package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
)

func dbConnection() (*sql.DB, error) {
	// Define database connection settings
	dsn := os.Getenv("MYSQL_USER") + ":" + os.Getenv("MYSQL_PASSWORD") + "@tcp(database:3306)/db"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// Verify connection with a ping
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// executeNonQuery wraps the db.ExecContext call to reduce duplication and centralize error handling.
func executeNonQuery(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (int64, error) {
	log.Printf("Executing query: %s", query)
	if query == "" {
		return 0, fmt.Errorf("query string is empty")
	}

	res, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %v, error: %w", query, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID for query: %v, error: %w", query, err)
	}
	return id, nil
}

func SelectNames(db *sql.DB, ctx context.Context, query string) ([]Name, error) {
	var names []Name // [[firstName, lastName], [firstName, lastName]]

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var name Name
		rows.Scan(&name.FirstName, &name.LastName)
		names = append(names, name)
	}

	return names, nil
}

func SelectPerson(db *sql.DB, ctx context.Context, query string, firstName, lastName string) (int, Person, error) {
	var person Person
	var personId int

	err := db.QueryRowContext(ctx, query, firstName, lastName).Scan(
		&personId,
		&person.FirstName,
		&person.LastName,
		&person.Location,
		&person.Github,
		&person.Linkedin,
		&person.Summary,
		&person.Proficiencies,
		&person.Familiarities,
	)

	log.Print("Person selected successfully")

	return personId, person, err
}

func InsertPerson(db *sql.DB, ctx context.Context, query string, person Person) (int, error) {
	res, err := db.Exec(
		query,
		person.FirstName,
		person.LastName,
		person.Location,
		person.Github,
		person.Linkedin,
		person.Summary,
		person.Proficiencies,
		person.Familiarities,
	)
	if err != nil {
		return -1, err
	}

	personId, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	log.Print("Person inserted successfully")

	return int(personId), nil
}

func SelectEducation(db *sql.DB, ctx context.Context, query string, personId int) ([]Education, error) {
	var education []Education

	rows, err := db.QueryContext(ctx, query, personId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var degree Education

		if err = rows.Scan(
			&degree.Institution,
			&degree.Degree,
			&degree.Duration,
			&degree.Gpa,
			&degree.SpecialNotes,
		); err != nil {
			return nil, err
		}

		education = append(education, degree)
	}

	log.Print("Education selected successfully")

	return education, nil
}

func InsertEducation(db *sql.DB, ctx context.Context, insertStmt, deleteStmt string, education []Education, personId int) error {
	_, err := db.Exec(deleteStmt, personId) // Deleting the existing records to allow for a clean update
	if err != nil {
		return err
	}

	for _, degree := range education {
		_, err = db.Exec(
			insertStmt,
			personId,
			degree.Institution,
			degree.Degree,
			degree.Duration,
			degree.Gpa,
			degree.SpecialNotes,
		)
		if err != nil {
			return err
		}
	}

	log.Print("Education inserted successfully")

	return nil
}

func SelectJobs(db *sql.DB, ctx context.Context, query string, personId int) ([]Job, error) {
	var jobs []Job
	jobMap := make(map[int]Job)

	rows, err := db.QueryContext(ctx, query, personId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id           int
			employer     string
			location     string
			duration     string
			title        string
			technologies string
			experience   string
		)

		if err = rows.Scan(
			&id,
			&employer,
			&location,
			&duration,
			&title,
			&technologies,
			&experience,
		); err != nil {
			return nil, err
		}

		job, exists := jobMap[id]
		if !exists {
			job = Job{
				Employer:     employer,
				Location:     location,
				Duration:     duration,
				Title:        title,
				Technologies: technologies,
				Experiences:  []string{},
			}
		}

		job.Experiences = append(job.Experiences, experience)
		jobMap[id] = job
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	for _, job := range jobMap {
		jobs = append(jobs, job)
	}

	log.Print("Jobs selected successfully")

	return jobs, nil
}

func InsertJobs(
	db *sql.DB,
	ctx context.Context,
	insertEmployerStmt string,
	insertJobStmt string,
	insertExperienceStmt string,
	selectJobIdsQuery string,
	deleteJobsStmt string,
	deleteJobExperiencesStmt string,
	jobs []Job,
	personId int,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err = deleteExistingJobsAndExperiences(ctx, tx, personId, selectJobIdsQuery, deleteJobsStmt, deleteJobExperiencesStmt); err != nil {
		return err
	}

	for _, job := range jobs {
		// Insert employer and get the ID
		employerId, err := executeNonQuery(ctx, tx, insertEmployerStmt, job.Employer, job.Location)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert employer: %w", err)
		}

		// Insert job and get the ID
		jobId, err := executeNonQuery(ctx, tx, insertJobStmt, personId, employerId, job.Duration, job.Title, job.Technologies)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert job: %w", err)
		}

		// Insert job_experiences
		for _, experience := range job.Experiences {
			_, err := executeNonQuery(ctx, tx, insertExperienceStmt, jobId, experience)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to insert job_experience: %w", err)
			}
		}
	}

	// Commit the transaction after all operations are successful
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Print("Jobs inserted successfuly")

	return nil
}

// deleteExistingJobsAndExperiences deletes all the job and job_experience entries which correspond with the personId.
func deleteExistingJobsAndExperiences(ctx context.Context, tx *sql.Tx, personId int, selectJobIdsQuery, deleteJobsStmt, deleteJobExperiencesStmt string) error {
	rows, err := db.QueryContext(ctx, selectJobIdsQuery, personId)
	if err != nil {
		return fmt.Errorf("error retrieving job.id: %w", err)
	}

	for rows.Next() {
		var jobId int
		rows.Scan(&jobId)

		_, err = executeNonQuery(ctx, tx, deleteJobExperiencesStmt, jobId)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete job experiences: %w", err)
		}
	}

	_, err = executeNonQuery(ctx, tx, deleteJobsStmt, personId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete jobs: %w", err)
	}

	return nil
}

func SelectProjects(db *sql.DB, ctx context.Context, query string, personId int) ([]Project, error) {
	var projects []Project
	projectMap := make(map[int]Project)

	rows, err := db.QueryContext(ctx, query, personId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id           int
			name         string
			repo         string
			technologies string
			contribution string
		)

		if err = rows.Scan(
			&id,
			&name,
			&repo,
			&technologies,
			&contribution,
		); err != nil {
			return nil, err
		}

		project, exists := projectMap[id]
		if !exists {
			project = Project{
				Name:          name,
				Repository:    repo,
				Technologies:  technologies,
				Contributions: []string{},
			}
		}
		project.Contributions = append(project.Contributions, contribution)
		projectMap[id] = project
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	for _, project := range projectMap {
		projects = append(projects, project)
	}

	log.Print("Projects selected successfully")
	return projects, nil
}

func InsertProjects(db *sql.DB, ctx context.Context, insertProjectStmt, insertContributionStmt, selectProjectsQuery, deleteProjectsStmt, deleteContributionsStmt string, projects []Project, personId int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err = deleteExistingProjectsAndContributions(ctx, tx, personId, selectProjectsQuery, deleteProjectsStmt, deleteContributionsStmt); err != nil {
		return err
	}

	for _, project := range projects {
		// Insert into project
		projectId, err := executeNonQuery(
			ctx,
			tx,
			insertProjectStmt,
			personId,
			project.Name,
			project.Repository,
			project.Technologies,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert project: %w", err)
		}

		// Insert into project_contribution
		for _, contribution := range project.Contributions {
			_, err := executeNonQuery(ctx, tx, insertContributionStmt, projectId, contribution)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to insert project_contribution: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Print("Projects inserted successfully")
	return nil
}

// deleteExistingProjectsAndContributions deletes all the project and project_contribution entries which correspond with the personId.
func deleteExistingProjectsAndContributions(ctx context.Context, tx *sql.Tx, personId int, selectProjectsQuery, deleteProjectsStmt, deleteContributionsStmt string) error {
	rows, err := db.QueryContext(ctx, selectProjectsQuery, personId)
	if err != nil {
		return fmt.Errorf("failed to retrieve project.id: %w", err)
	}

	for rows.Next() {
		var projectId int
		rows.Scan(&projectId)

		_, err = executeNonQuery(ctx, tx, deleteProjectsStmt, projectId)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete contributions: %w", err)
		}
	}

	_, err = executeNonQuery(ctx, tx, deleteContributionsStmt, personId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete projects: %w", err)
	}

	return nil
}

func SelectCertifications(db *sql.DB, ctx context.Context, query string, personId int) ([]Certification, error) {
	var certifications []Certification

	rows, err := db.QueryContext(ctx, query, personId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cert Certification

		if err = rows.Scan(
			&cert.Organization,
			&cert.Certification,
			&cert.Expiration,
		); err != nil {
			return nil, err
		}

		certifications = append(certifications, cert)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	log.Print("Certification selected successfully")
	return certifications, nil
}

func InsertCertifications(db *sql.DB, ctx context.Context, insertOrgStmt, insertCertStmt, deleteCertsStmt string, certs []Certification, personId int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = executeNonQuery(ctx, tx, deleteCertsStmt, personId) // Deleting the existing records to allow for clean update
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete certification: %w", err)
	}

	for _, cert := range certs {
		// Insert into certifying_org
		orgId, err := executeNonQuery(ctx, tx, insertOrgStmt, cert.Organization)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert certifying_org: %w", err)
		}

		// Insert into certification
		_, err = executeNonQuery(ctx, tx, insertCertStmt, int(orgId), personId, cert.Certification, cert.Expiration)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert certification: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Print("Certification inserted successfully")
	return nil
}
