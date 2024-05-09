CREATE TABLE person (
    id INT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    location VARCHAR(255),
    portfolio_link VARCHAR(255),
    linkedin_link VARCHAR(255),
    professional_summary TEXT,
    proficient_skills TEXT,
    familiar_skills TEXT,
    UNIQUE KEY (portfolio_link)
);

CREATE TABLE education (
    education_id INT AUTO_INCREMENT PRIMARY KEY,
    person_id INT,
    institution VARCHAR(255),
    degree VARCHAR(255),
    duration VARCHAR(19),
    gpa VARCHAR(9),
    special_notes TEXT,
    FOREIGN KEY (person_id) REFERENCES person(id),
    UNIQUE KEY (person_id, institution, degree)
);

CREATE TABLE employer (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255),
    location VARCHAR(255),
    UNIQUE KEY (name, location)
);

CREATE TABLE job (
    id INT AUTO_INCREMENT PRIMARY KEY,
    person_id INT,
    employer_id INT,
    duration VARCHAR(19),
    title VARCHAR(255),
    technologies TEXT,
    FOREIGN KEY (employer_id) REFERENCES employer(id),
    FOREIGN KEY (person_id) REFERENCES person(id),
    UNIQUE KEY (employer_id, title)
);

CREATE TABLE job_experience (
    id INT AUTO_INCREMENT PRIMARY KEY,
    job_id INT,
    experience TEXT,
    FOREIGN KEY (job_id) REFERENCES job(id),
    UNIQUE KEY (job_id, experience(255))
);

CREATE TABLE project (
    id INT AUTO_INCREMENT PRIMARY KEY,
    person_id INT,
    name VARCHAR(255),
    repository_link VARCHAR(255),
    technologies TEXT,
    FOREIGN KEY (person_id) REFERENCES person(id),
    UNIQUE KEY (person_id, name, repository_link)
);

CREATE TABLE project_contribution (
    id INT AUTO_INCREMENT PRIMARY KEY,
    project_id INT,
    contribution TEXT,
    FOREIGN KEY (project_id) REFERENCES project(id),
    UNIQUE KEY (project_id, contribution(255))
);

CREATE TABLE certifying_organization (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255),
    UNIQUE KEY (name)
);

CREATE TABLE certification (
    id INT AUTO_INCREMENT PRIMARY KEY,
    organization_id INT,
    person_id INT,
    name VARCHAR(255),
    expiration_date CHAR(8),
    FOREIGN KEY (organization_id) REFERENCES certifying_organization(id),
    UNIQUE KEY (organization_id, person_id, name)
);
