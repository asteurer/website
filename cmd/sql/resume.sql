--name: SelectPerson
SELECT
    person.id,
    person.first_name,
    person.last_name,
    person.location,
    person.portfolio_link,
    person.linkedin_link,
    person.professional_summary,
    person.proficient_skills,
    person.familiar_skills
FROM person 
WHERE
    LOWER(person.first_name)  = LOWER(?)
    AND LOWER(person.last_name) = LOWER(?)
LIMIT 1;

--name: SelectNames
SELECT first_name, last_name FROM person;

--name: InsertPerson
INSERT INTO person (
    first_name,
    last_name,
    location,
    portfolio_link,
    linkedin_link,
    professional_summary,
    proficient_skills,
    familiar_skills
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    id = LAST_INSERT_ID(id),
    first_name = VALUES(first_name),
    last_name = VALUES(last_name),
    location = VALUES(location),
    portfolio_link = VALUES(portfolio_link),
    linkedin_link = VALUES(linkedin_link),
    professional_summary = VALUES(professional_summary),
    proficient_skills = VALUES(proficient_skills),
    familiar_skills = VALUES(familiar_skills);

--name: SelectEducation
SELECT
    institution,
    degree,
    duration,
    gpa,
    special_notes
FROM education
WHERE person_id = ?;

--name: InsertEducation
INSERT INTO education (
    person_id,
    institution,
    degree,
    duration,
    gpa,
    special_notes
)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    institution = VALUES(institution),
    degree = VALUES(degree),
    duration = VALUES(duration),
    gpa = VALUES(gpa),
    special_notes = VALUES(special_notes);


--name: SelectJobs
SELECT
    job.id,
    employer.name,
    employer.location,
    job.duration,
    job.title,
    job.technologies,
    job_experience.experience
FROM job_experience
LEFT JOIN job
    ON job.id = job_experience.job_id
LEFT JOIN employer
    ON employer.id = job.employer_id
WHERE job.person_id = ?;

--name: InsertEmployer
INSERT INTO employer (name, location) 
VALUES (?, ?)
ON DUPLICATE KEY UPDATE 
    id = LAST_INSERT_ID(id),
    name = VALUES(name),
    location = VALUES(location);

--name: InsertJob
INSERT INTO job (
    person_id, 
    employer_id, 
    duration, 
    title, 
    technologies
)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    id = LAST_INSERT_ID(id),
    duration = VALUES(duration),
    title = VALUES(title),
    technologies = VALUES(technologies);

--name: InsertJobExperience
INSERT INTO job_experience (job_id, experience) 
VALUES (?, ?)
ON DUPLICATE KEY UPDATE
    experience = VALUES(experience); 

--name: SelectProjects
SELECT
    project.id,
    project.name,
    project.repository_link,
    project.technologies,
    project_contribution.contribution
FROM project_contribution
LEFT JOIN project
    ON project.id = project_contribution.project_id
WHERE project.person_id = ?;

--name: InsertProject
INSERT INTO project (
    person_id,
    name,
    repository_link,
    technologies
)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    id = LAST_INSERT_ID(id),
    person_id = VALUES(person_id),
    name = VALUES(name),
    repository_link = VALUES(repository_link),
    technologies = VALUES(technologies);

--name: InsertProjectContribution
INSERT INTO project_contribution (project_id, contribution) 
VALUES (?, ?)
ON DUPLICATE KEY UPDATE
    project_id = VALUES(project_id),
    contribution = VALUES(contribution);

--name: SelectCertifications
SELECT
    certifying_organization.name,
    certification.name,
    certification.expiration_date
FROM certification
LEFT JOIN certifying_organization
    ON certifying_organization.id = certification.organization_id
WHERE certification.person_id = ?;

--name: InsertCertifyingOrg
INSERT INTO certifying_organization (name) 
VALUES (?)
ON DUPLICATE KEY UPDATE
    id = LAST_INSERT_ID(id),
    name = VALUES(name);

--name: InsertCertification
INSERT INTO certification (
    organization_id, 
    person_id, 
    name, 
    expiration_date
)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    organization_id = VALUES(organization_id),
    person_id = VALUES(person_id),
    name = VALUES(name),
    expiration_date = VALUES(expiration_date);