package app

import (
	"strings"

	"profzom/internal/domain/profile"
)

func StudentCompletion(p profile.StudentProfile) int {
	total := 6
	filled := 0
	if strings.TrimSpace(p.Name) != "" {
		filled++
	}
	if strings.TrimSpace(p.University) != "" {
		filled++
	}
	if p.Course > 0 {
		filled++
	}
	if strings.TrimSpace(p.Specialty) != "" {
		filled++
	}
	if len(p.Skills) > 0 {
		filled++
	}
	if strings.TrimSpace(p.About) != "" {
		filled++
	}
	return int(float64(filled) / float64(total) * 100)
}

func CompanyCompletion(p profile.CompanyProfile) int {
	total := 4
	filled := 0
	if strings.TrimSpace(p.Name) != "" {
		filled++
	}
	if strings.TrimSpace(p.Industry) != "" {
		filled++
	}
	if strings.TrimSpace(p.Description) != "" {
		filled++
	}
	if strings.TrimSpace(p.ContactName) != "" && HasCompanyContact(p) {
		filled++
	}
	return int(float64(filled) / float64(total) * 100)
}

func IsStudentProfileComplete(p profile.StudentProfile) bool {
	return StudentCompletion(p) == 100
}

func IsCompanyProfileComplete(p profile.CompanyProfile) bool {
	return CompanyCompletion(p) == 100
}

func HasCompanyContact(p profile.CompanyProfile) bool {
	return strings.TrimSpace(p.ContactEmail) != "" || strings.TrimSpace(p.ContactPhone) != ""
}
