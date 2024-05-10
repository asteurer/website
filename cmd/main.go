package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type Person struct {
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Location      string `json:"location"`
	Github        string `json:"github"`
	Linkedin      string `json:"linkedin"`
	Summary       string `json:"summary"`
	Proficiencies string `json:"proficiencies"`
	Familiarities string `json:"familiarities"`
}

type Education struct {
	Institution  string `json:"institution"`
	Degree       string `json:"degree"`
	Duration     string `json:"duration"`
	Gpa          string `json:"gpa"`
	SpecialNotes string `json:"specialNotes"`
}

type Job struct {
	Employer     string   `json:"employer"`
	Location     string   `json:"location"`
	Title        string   `json:"title"`
	Duration     string   `json:"duration"`
	Technologies string   `json:"technologies"`
	Experiences  []string `json:"experiences"`
}

type Project struct {
	Name          string   `json:"name"`
	Repository    string   `json:"repository"`
	Technologies  string   `json:"technologies"`
	Contributions []string `json:"contributions"`
}

type Certification struct {
	Organization  string `json:"organization"`
	Certification string `json:"certification"`
	Expiration    string `json:"expiration"`
}

type Resume struct {
	Person         Person          `json:"person"`
	Education      []Education     `json:"education"`
	WorkExperience []Job           `json:"work_experience"`
	Projects       []Project       `json:"projects"`
	Certifications []Certification `json:"certifications"`
}

type Name struct {
	FirstName string
	LastName  string
}

var db *sql.DB
var sqlQueries = make(map[string]map[string]string)

func main() {
	router := gin.Default()
	gin.SetMode(gin.ReleaseMode) // Set release mode

	var err error
	db, err = dbConnection()
	if err != nil {
		log.Fatalf("error setting up database: %v", err)
	}
	defer db.Close()

	if err := LoadSQLFiles("./sql"); err != nil {
		log.Fatal("SQL files failed to load:", err)
	}

	router.LoadHTMLGlob("templates/*")
	router.Static("/media", "./media")
	router.GET("/", getIndexHTML)
	router.GET("/resume/:firstname/:lastname", getResumeHTML)
	router.GET("/api/resume/:firstname/:lastname", getResumeJSON)
	router.POST("/api/resume/", placeResumeJSON)

	// Start the server only if the database connection is successful
	if err := router.Run("0.0.0.0:8080"); err != nil {
		log.Fatal("server failed to start:", err)
	}
}

func fetchData(firstname, lastname string, ctx context.Context) (Resume, error) {
	var resume Resume

	personId, person, err := SelectPerson(db, ctx, sqlQueries["./sql/resume.sql"]["SelectPerson"], firstname, lastname)
	if err != nil {
		return resume, err
	}

	education, err := SelectEducation(db, ctx, sqlQueries["./sql/resume.sql"]["SelectEducation"], personId)
	if err != nil {
		return resume, err
	}

	jobs, err := SelectJobs(db, ctx, sqlQueries["./sql/resume.sql"]["SelectJobs"], personId)
	if err != nil {
		return resume, err
	}

	projects, err := SelectProjects(db, ctx, sqlQueries["./sql/resume.sql"]["SelectProjects"], personId)
	if err != nil {
		return resume, err
	}

	certs, err := SelectCertifications(db, ctx, sqlQueries["./sql/resume.sql"]["SelectCertifications"], personId)
	if err != nil {
		return resume, err
	}

	resume.Person = person
	resume.Education = append(resume.Education, education...)
	resume.WorkExperience = append(resume.WorkExperience, jobs...)
	resume.Projects = append(resume.Projects, projects...)
	resume.Certifications = append(resume.Certifications, certs...)

	return resume, nil
}

func getResumeJSON(c *gin.Context) {
	firstname := c.Param("firstname")
	lastname := c.Param("lastname")
	context := c.Request.Context()

	resume, err := fetchData(firstname, lastname, context)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "No resume found"})
		} else {
			log.Printf("Error querying resume: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, resume)
}

func getResumeHTML(c *gin.Context) {
	firstname := c.Param("firstname")
	lastname := c.Param("lastname")
	ctx := c.Request.Context()

	resume, err := fetchData(firstname, lastname, ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "No resume found"})
		} else {
			log.Printf("Error querying table: %v\n", err)
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Error querying resume"})
		}
		return
	}

	c.HTML(200, "resume.tmpl", gin.H{
		"Resume": resume,
	})
}

func getIndexHTML(c *gin.Context) {
	names, err := SelectNames(db, c.Request.Context(), sqlQueries["./sql/resume.sql"]["SelectNames"])
	if err != nil {
		if err == sql.ErrNoRows {
			c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "No people found"})
		} else {
			log.Printf("Error querying table: %v\n", err)
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Error querying resume names"})
		}
		return
	}

	c.HTML(200, "index.tmpl", gin.H{
		"Names": names,
	})
}

func placeResumeJSON(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization token not provided"})
	}

	isValid, err := validateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	if !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	var resumes []Resume

	if err := c.ShouldBindJSON(&resumes); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, resume := range resumes {
		person := resume.Person
		education := resume.Education
		jobs := resume.WorkExperience
		projects := resume.Projects
		certs := resume.Certifications

		personId, err := InsertPerson(db, c.Request.Context(), sqlQueries["./sql/resume.sql"]["InsertPerson"], person)
		if err != nil {
			log.Printf("SQL error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err = InsertEducation(
			db,
			c.Request.Context(),
			sqlQueries["./sql/resume.sql"]["InsertEducation"],
			sqlQueries["./sql/resume.sql"]["DeleteEducation"],
			education,
			personId,
		); err != nil {
			log.Printf("SQL error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err = InsertJobs(
			db,
			c.Request.Context(),
			sqlQueries["./sql/resume.sql"]["InsertEmployer"],
			sqlQueries["./sql/resume.sql"]["InsertJob"],
			sqlQueries["./sql/resume.sql"]["InsertJobExperience"],
			sqlQueries["./sql/resume.sql"]["SelectJobIds"],
			sqlQueries["./sql/resume.sql"]["DeleteJobs"],
			sqlQueries["./sql/resume.sql"]["DeleteJobExperiences"],
			jobs,
			personId,
		); err != nil {
			log.Printf("SQL error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err = InsertProjects(
			db,
			c.Request.Context(),
			sqlQueries["./sql/resume.sql"]["InsertProject"],
			sqlQueries["./sql/resume.sql"]["InsertProjectContribution"],
			sqlQueries["./sql/resume.sql"]["SelectProjectIds"],
			sqlQueries["./sql/resume.sql"]["DeleteProjects"],
			sqlQueries["./sql/resume.sql"]["DeleteProjectContributions"],
			projects,
			personId,
		); err != nil {
			log.Printf("SQL error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err = InsertCertifications(
			db,
			c.Request.Context(),
			sqlQueries["./sql/resume.sql"]["InsertCertifyingOrg"],
			sqlQueries["./sql/resume.sql"]["InsertCertification"],
			sqlQueries["./sql/resume.sql"]["DeleteCertifications"],
			certs,
			personId,
		); err != nil {
			log.Printf("SQL error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Println("All resumes processed successfully")
		c.JSON(http.StatusAccepted, gin.H{"success": "resume has been posted"})
	}
}
