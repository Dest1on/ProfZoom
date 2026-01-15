package handlers

import (
    "net/http"
    "net/mail"
    "regexp"
    "strings"

    "profzom/internal/app"
    "profzom/internal/common"
    "profzom/internal/domain/profile"
    "profzom/internal/http/middleware"
    "profzom/internal/http/response"
)

type ProfileHandler struct {
    profiles *app.ProfileService
}

func NewProfileHandler(profiles *app.ProfileService) *ProfileHandler {
    return &ProfileHandler{profiles: profiles}
}

type studentProfileRequest struct {
    Name       string   `json:"name"`
    University string   `json:"university"`
    Course     int      `json:"course"`
    Specialty  string   `json:"specialty"`
    Skills     []string `json:"skills"`
    About      string   `json:"about"`
}

type companyProfileRequest struct {
    Name         string `json:"name"`
    Industry     string `json:"industry"`
    Description  string `json:"description"`
    ContactName  string `json:"contact_name"`
    ContactEmail string `json:"contact_email"`
    ContactPhone string `json:"contact_phone"`
}

func (h *ProfileHandler) GetStudent(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.UserIDFromContext(r.Context())
    if !ok {
        response.Error(w, errUnauthorized())
        return
    }
    profile, err := h.profiles.GetStudent(r.Context(), userID)
    if err != nil {
        response.Error(w, err)
        return
    }
    response.JSON(w, http.StatusOK, map[string]interface{}{
        "profile":    profile,
        "completion": app.StudentCompletion(*profile),
    })
}

func (h *ProfileHandler) UpsertStudent(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.UserIDFromContext(r.Context())
    if !ok {
        response.Error(w, errUnauthorized())
        return
    }
    var req studentProfileRequest
    if err := decodeJSON(r, &req); err != nil {
        response.Error(w, err)
        return
    }
    if err := validateStudentProfile(req); err != nil {
        response.Error(w, err)
        return
    }
    updated, err := h.profiles.UpsertStudent(r.Context(), profile.StudentProfile{
        UserID:     userID,
        Name:       req.Name,
        University: req.University,
        Course:     req.Course,
        Specialty:  req.Specialty,
        Skills:     req.Skills,
        About:      req.About,
    })
    if err != nil {
        response.Error(w, err)
        return
    }
    response.JSON(w, http.StatusOK, map[string]interface{}{
        "profile":    updated,
        "completion": app.StudentCompletion(*updated),
    })
}

func (h *ProfileHandler) GetCompany(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.UserIDFromContext(r.Context())
    if !ok {
        response.Error(w, errUnauthorized())
        return
    }
    profile, err := h.profiles.GetCompany(r.Context(), userID)
    if err != nil {
        response.Error(w, err)
        return
    }
    response.JSON(w, http.StatusOK, map[string]interface{}{
        "profile":    profile,
        "completion": app.CompanyCompletion(*profile),
    })
}

func (h *ProfileHandler) UpsertCompany(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.UserIDFromContext(r.Context())
    if !ok {
        response.Error(w, errUnauthorized())
        return
    }
    var req companyProfileRequest
    if err := decodeJSON(r, &req); err != nil {
        response.Error(w, err)
        return
    }
    if err := validateCompanyProfile(req); err != nil {
        response.Error(w, err)
        return
    }
    updated, err := h.profiles.UpsertCompany(r.Context(), profile.CompanyProfile{
        UserID:       userID,
        Name:         req.Name,
        Industry:     req.Industry,
        Description:  req.Description,
        ContactName:  req.ContactName,
        ContactEmail: req.ContactEmail,
        ContactPhone: req.ContactPhone,
    })
    if err != nil {
        response.Error(w, err)
        return
    }
    response.JSON(w, http.StatusOK, map[string]interface{}{
        "profile":    updated,
        "completion": app.CompanyCompletion(*updated),
    })
}

func validateStudentProfile(req studentProfileRequest) error {
    fields := map[string]string{}
    if strings.TrimSpace(req.Name) == "" {
        fields["name"] = "name is required"
    }
    if strings.TrimSpace(req.University) == "" {
        fields["university"] = "university is required"
    }
    if req.Course <= 0 {
        fields["course"] = "course must be > 0"
    }
    if strings.TrimSpace(req.Specialty) == "" {
        fields["specialty"] = "specialty is required"
    }
    if len(req.Skills) == 0 {
        fields["skills"] = "at least one skill is required"
    }
    if strings.TrimSpace(req.About) == "" {
        fields["about"] = "about is required"
    }
    if len(fields) > 0 {
        return common.NewValidationError("invalid student profile", fields)
    }
    return nil
}

func validateCompanyProfile(req companyProfileRequest) error {
    fields := map[string]string{}
    if strings.TrimSpace(req.Name) == "" {
        fields["name"] = "name is required"
    }
    if strings.TrimSpace(req.Industry) == "" {
        fields["industry"] = "industry is required"
    }
    if strings.TrimSpace(req.Description) == "" {
        fields["description"] = "description is required"
    }
    if strings.TrimSpace(req.ContactName) == "" {
        fields["contact_name"] = "contact_name is required"
    }
    if strings.TrimSpace(req.ContactEmail) == "" && strings.TrimSpace(req.ContactPhone) == "" {
        fields["contact"] = "provide at least one contact: email or phone"
    }
    if req.ContactEmail != "" {
        if _, err := mail.ParseAddress(req.ContactEmail); err != nil {
            fields["contact_email"] = "invalid email format"
        }
    }
    if req.ContactPhone != "" {
        if !regexp.MustCompile(`^\+?[0-9]{7,15}$`).MatchString(req.ContactPhone) {
            fields["contact_phone"] = "invalid phone format"
        }
    }
    if len(fields) > 0 {
        return common.NewValidationError("invalid company profile", fields)
    }
    return nil
}
