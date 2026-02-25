package handler

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"suv/internal/config"
	"suv/internal/middleware"
	"suv/internal/model"
)

type Handler struct {
	db  *sql.DB
	cfg *config.Config
}

func New(db *sql.DB, cfg *config.Config) *Handler {
	return &Handler{db: db, cfg: cfg}
}

// ==================== Pages ====================

func (h *Handler) IndexPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"csrf_token": getCSRF(c),
	})
}

func (h *Handler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"csrf_token": getCSRF(c),
	})
}

func (h *Handler) RecordPage(c *gin.Context) {
	user := getUser(c)
	c.HTML(http.StatusOK, "record.html", gin.H{
		"user":       user,
		"csrf_token": getCSRF(c),
	})
}

func (h *Handler) PublicPage(c *gin.Context) {
	c.HTML(http.StatusOK, "public.html", gin.H{
		"csrf_token": getCSRF(c),
	})
}

func (h *Handler) AuditPage(c *gin.Context) {
	user := getUser(c)
	c.HTML(http.StatusOK, "audit.html", gin.H{
		"user":       user,
		"csrf_token": getCSRF(c),
	})
}

func (h *Handler) ExportPage(c *gin.Context) {
	user := getUser(c)
	c.HTML(http.StatusOK, "export.html", gin.H{
		"user":       user,
		"csrf_token": getCSRF(c),
	})
}

// ==================== Auth API ====================

func (h *Handler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入用户名和密码"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名或密码不能为空"})
		return
	}

	var user model.User
	err := h.db.QueryRow(
		"SELECT id, username, password_hash, display_name, role FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.DisplayName, &user.Role)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	token, err := middleware.GenerateToken(h.cfg.JWTSecret, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误"})
		return
	}

	c.SetCookie("token", token, 86400, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"role":         user.Role,
		},
	})
}

func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "已注销"})
}

func (h *Handler) GetCurrentUser(c *gin.Context) {
	user := getUser(c)
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.UserID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// ==================== Violations API ====================

func (h *Handler) CreateViolation(c *gin.Context) {
	user := getUser(c)

	var req model.ViolationRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写完整信息: " + err.Error()})
		return
	}

	// Handle photo upload
	photoPath := ""
	file, header, err := c.Request.FormFile("photo")
	if err == nil {
		defer file.Close()

		// Validate file size
		if header.Size > h.cfg.MaxUpload {
			c.JSON(http.StatusBadRequest, gin.H{"error": "照片大小不能超过 5MB"})
			return
		}

		// Validate file type
		ext := strings.ToLower(filepath.Ext(header.Filename))
		allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
		if !allowed[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 JPG/PNG/GIF/WebP 格式的图片"})
			return
		}

		// Read first 512 bytes to detect MIME type
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		mimeType := http.DetectContentType(buf[:n])
		if !strings.HasPrefix(mimeType, "image/") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "文件类型不合法"})
			return
		}
		file.Seek(0, 0)

		// Save file
		filename := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), user.UserID, ext)
		savePath := filepath.Join(h.cfg.UploadDir, filename)

		os.MkdirAll(h.cfg.UploadDir, 0755)

		dst, err := os.Create(savePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败"})
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败"})
			return
		}

		photoPath = filename
	}

	result, err := h.db.Exec(
		`INSERT INTO violations (dorm, student_name, class_name, period, reason, department, inspector, photo_path, created_by)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.Dorm, req.StudentName, req.ClassName, req.Period, req.Reason, req.Department, req.Inspector, photoPath, user.UserID,
	)
	if err != nil {
		log.Printf("Insert violation error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "提交成功"})
}

func (h *Handler) ListViolations(c *gin.Context) {
	dateStr := c.Query("date")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	offset := (page - 1) * limit

	where := "WHERE 1=1"
	args := []interface{}{}

	if dateStr != "" {
		where += " AND DATE(v.created_at) = ?"
		args = append(args, dateStr)
	}

	if keyword != "" {
		where += " AND (v.student_name LIKE ? OR v.class_name LIKE ? OR v.dorm LIKE ? OR v.reason LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw, kw)
	}

	// Count total
	var total int
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM violations v %s", where)
	h.db.QueryRow(countSQL, args...).Scan(&total)

	// Query data
	querySQL := fmt.Sprintf(`
		SELECT v.id, v.dorm, v.student_name, v.class_name, v.period, v.reason, 
		       v.department, v.inspector, v.photo_path, v.created_by, v.created_at,
		       COALESCE(u.display_name, u.username) as creator_name
		FROM violations v
		LEFT JOIN users u ON v.created_by = u.id
		%s
		ORDER BY v.created_at DESC
		LIMIT ? OFFSET ?
	`, where)

	queryArgs := append(args, limit, offset)
	rows, err := h.db.Query(querySQL, queryArgs...)
	if err != nil {
		log.Printf("Query violations error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	violations := []model.Violation{}
	for rows.Next() {
		var v model.Violation
		err := rows.Scan(&v.ID, &v.Dorm, &v.StudentName, &v.ClassName, &v.Period, &v.Reason,
			&v.Department, &v.Inspector, &v.PhotoPath, &v.CreatedBy, &v.CreatedAt, &v.CreatorName)
		if err != nil {
			continue
		}
		violations = append(violations, v)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  violations,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *Handler) GetTodayViolations(c *gin.Context) {
	today := time.Now().Format("2006-01-02")

	rows, err := h.db.Query(`
		SELECT v.id, v.dorm, v.student_name, v.class_name, v.period, v.reason,
		       v.department, v.inspector, v.photo_path, v.created_by, v.created_at,
		       COALESCE(u.display_name, u.username) as creator_name
		FROM violations v
		LEFT JOIN users u ON v.created_by = u.id
		WHERE DATE(v.created_at) = ?
		ORDER BY v.created_at DESC
	`, today)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	violations := []model.Violation{}
	for rows.Next() {
		var v model.Violation
		rows.Scan(&v.ID, &v.Dorm, &v.StudentName, &v.ClassName, &v.Period, &v.Reason,
			&v.Department, &v.Inspector, &v.PhotoPath, &v.CreatedBy, &v.CreatedAt, &v.CreatorName)
		violations = append(violations, v)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  violations,
		"date":  today,
		"count": len(violations),
	})
}

func (h *Handler) DeleteViolation(c *gin.Context) {
	id := c.Param("id")
	idNum, err := strconv.Atoi(id)
	if err != nil || idNum < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的记录 ID"})
		return
	}

	// Get photo path before deletion
	var photoPath string
	h.db.QueryRow("SELECT photo_path FROM violations WHERE id = ?", idNum).Scan(&photoPath)

	result, err := h.db.Exec("DELETE FROM violations WHERE id = ?", idNum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
		return
	}

	// Delete photo file
	if photoPath != "" {
		os.Remove(filepath.Join(h.cfg.UploadDir, photoPath))
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *Handler) GetViolationPhoto(c *gin.Context) {
	id := c.Param("id")
	idNum, err := strconv.Atoi(id)
	if err != nil || idNum < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的记录 ID"})
		return
	}

	var photoPath string
	err = h.db.QueryRow("SELECT photo_path FROM violations WHERE id = ?", idNum).Scan(&photoPath)
	if err != nil || photoPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "照片不存在"})
		return
	}

	fullPath := filepath.Join(h.cfg.UploadDir, photoPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "照片文件不存在"})
		return
	}

	c.File(fullPath)
}

// ==================== Export API ====================

func (h *Handler) ExportCSV(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	rows, err := h.db.Query(`
		SELECT v.id, v.dorm, v.student_name, v.class_name, v.period, v.reason,
		       v.department, v.inspector, v.created_at,
		       COALESCE(u.display_name, u.username) as creator_name
		FROM violations v
		LEFT JOIN users u ON v.created_by = u.id
		WHERE DATE(v.created_at) = ?
		ORDER BY v.created_at ASC
	`, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	// BOM for Excel UTF-8 compatibility
	bom := "\xEF\xBB\xBF"
	csv := bom + "ID,宿舍号,姓名,班级,时间段,违纪原因,部门,执勤人,记录时间,录入人\n"

	for rows.Next() {
		var id uint
		var dorm, name, class, period, reason, dept, inspector, creator string
		var createdAt time.Time
		rows.Scan(&id, &dorm, &name, &class, &period, &reason, &dept, &inspector, &createdAt, &creator)

		// Escape CSV fields
		reason = strings.ReplaceAll(reason, "\"", "\"\"")
		reason = strings.ReplaceAll(reason, "\n", " ")

		csv += fmt.Sprintf("%d,\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n",
			id, dorm, name, class, period, reason, dept, inspector,
			createdAt.Format("2006-01-02 15:04:05"), creator)
	}

	filename := fmt.Sprintf("violations_%s.csv", dateStr)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.String(http.StatusOK, csv)
}

// ==================== User Management (Admin) ====================

func (h *Handler) ListUsers(c *gin.Context) {
	rows, err := h.db.Query("SELECT id, username, display_name, role, created_at FROM users ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		var u model.User
		rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.Role, &u.CreatedAt)
		users = append(users, u)
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *Handler) CreateUser(c *gin.Context) {
	var body struct {
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required,min=6"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role" binding:"required,oneof=admin staff"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误"})
		return
	}

	if body.DisplayName == "" {
		body.DisplayName = body.Username
	}

	_, err = h.db.Exec(
		"INSERT INTO users (username, password_hash, display_name, role) VALUES (?, ?, ?, ?)",
		body.Username, string(hash), body.DisplayName, body.Role,
	)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户创建成功"})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	idNum, _ := strconv.Atoi(id)
	if idNum < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效 ID"})
		return
	}

	// Prevent deleting self
	user := getUser(c)
	if uint(idNum) == user.UserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除自己"})
		return
	}

	_, err := h.db.Exec("DELETE FROM users WHERE id = ?", idNum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码至少6位"})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	_, err := h.db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(hash), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "密码重置成功"})
}

// ==================== Stats ====================

func (h *Handler) GetStats(c *gin.Context) {
	today := time.Now().Format("2006-01-02")

	var todayCount, totalCount, userCount int
	h.db.QueryRow("SELECT COUNT(*) FROM violations WHERE DATE(created_at) = ?", today).Scan(&todayCount)
	h.db.QueryRow("SELECT COUNT(*) FROM violations").Scan(&totalCount)
	h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)

	c.JSON(http.StatusOK, gin.H{
		"today_count": todayCount,
		"total_count": totalCount,
		"user_count":  userCount,
	})
}

// ==================== Seed default admin ====================

func (h *Handler) SeedAdmin() {
	var count int
	h.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if count > 0 {
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	_, err := h.db.Exec(
		"INSERT INTO users (username, password_hash, display_name, role) VALUES (?, ?, ?, ?)",
		"admin", string(hash), "系统管理员", "admin",
	)
	if err != nil {
		log.Printf("Seed admin error: %v", err)
		return
	}
	log.Println("Default admin created: admin / admin123")
}

// ==================== Template Functions ====================

func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
}

// ==================== Helpers ====================

func getUser(c *gin.Context) model.Claims {
	user, _ := c.Get("user")
	return user.(model.Claims)
}

func getCSRF(c *gin.Context) string {
	token, _ := c.Get("csrf_token")
	if token == nil {
		return ""
	}
	return token.(string)
}
