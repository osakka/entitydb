package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	baseURL = "http://localhost:8085"
	adminUser = "admin"
	adminPass = "admin"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type CreateEntityRequest struct {
	Tags    []string          `json:"tags"`
	Content map[string]interface{} `json:"content"`
}

func main() {
	rand.Seed(time.Now().UnixNano())
	
	// Login first
	token, err := login()
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	fmt.Println("✅ Logged in successfully")
	
	// Create organizations
	fmt.Println("\n📋 Creating organizations...")
	orgIDs := createOrganizations(token)
	
	// Create projects
	fmt.Println("\n📁 Creating projects...")
	projectIDs := createProjects(token, orgIDs)
	
	// Create epics
	fmt.Println("\n🎯 Creating epics...")
	epicIDs := createEpics(token, projectIDs)
	
	// Create stories
	fmt.Println("\n📖 Creating stories...")
	storyIDs := createStories(token, epicIDs, projectIDs)
	
	// Create team members
	fmt.Println("\n👥 Creating team members...")
	userIDs := createTeamMembers(token)
	
	// Create tasks with various statuses
	fmt.Println("\n📋 Creating tasks...")
	createTasks(token, storyIDs, userIDs, projectIDs)
	
	// Create some historical tasks for metrics
	fmt.Println("\n📊 Creating historical data...")
	createHistoricalTasks(token, storyIDs, userIDs, projectIDs)
	
	fmt.Println("\n✅ Test data population complete!")
	fmt.Println("📈 Metrics should now be visible in the Reports view")
}

func login() (string, error) {
	loginReq := LoginRequest{
		Username: adminUser,
		Password: adminPass,
	}
	
	data, _ := json.Marshal(loginReq)
	resp, err := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", err
	}
	
	return loginResp.Token, nil
}

func createEntity(token string, tags []string, content map[string]interface{}) (string, error) {
	req := CreateEntityRequest{
		Tags:    tags,
		Content: content,
	}
	
	data, _ := json.Marshal(req)
	
	httpReq, err := http.NewRequest("POST", baseURL+"/api/v1/entities/create", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	if id, ok := result["id"].(string); ok {
		return id, nil
	}
	
	return "", fmt.Errorf("no ID in response")
}

func createOrganizations(token string) []string {
	orgs := []struct {
		name string
		desc string
	}{
		{"TechCorp Solutions", "Leading technology solutions provider"},
		{"Digital Innovations", "Cutting-edge digital transformation"},
		{"CloudFirst Inc", "Cloud infrastructure specialists"},
	}
	
	var ids []string
	for _, org := range orgs {
		tags := []string{
			"hub:worca",
			"type:organization",
			"status:active",
			"name:" + org.name,
		}
		content := map[string]interface{}{
			"name":        org.name,
			"description": org.desc,
			"industry":    "technology",
			"size":        "medium",
		}
		
		id, err := createEntity(token, tags, content)
		if err != nil {
			log.Printf("Failed to create org %s: %v", org.name, err)
			continue
		}
		ids = append(ids, id)
		fmt.Printf("  ✓ Created organization: %s\n", org.name)
	}
	
	return ids
}

func createProjects(token string, orgIDs []string) []string {
	projects := []struct {
		name string
		desc string
	}{
		{"Mobile Banking App", "Next-generation mobile banking"},
		{"Customer Portal", "Web-based customer service portal"},
		{"Data Analytics Platform", "Real-time analytics dashboard"},
		{"Security Framework", "Enterprise security solution"},
		{"API Gateway", "Microservices API management"},
	}
	
	var ids []string
	for i, proj := range projects {
		orgID := orgIDs[i%len(orgIDs)]
		tags := []string{
			"hub:worca",
			"type:project",
			"status:active",
			"name:" + proj.name,
			"organization:" + orgID,
		}
		content := map[string]interface{}{
			"name":        proj.name,
			"description": proj.desc,
			"start_date":  time.Now().AddDate(0, -3, 0).Format(time.RFC3339),
			"team_size":   rand.Intn(10) + 5,
		}
		
		id, err := createEntity(token, tags, content)
		if err != nil {
			log.Printf("Failed to create project %s: %v", proj.name, err)
			continue
		}
		ids = append(ids, id)
		fmt.Printf("  ✓ Created project: %s\n", proj.name)
	}
	
	return ids
}

func createEpics(token string, projectIDs []string) []string {
	epics := []struct {
		name string
		desc string
	}{
		{"User Authentication", "Complete authentication system"},
		{"Payment Integration", "Payment processing system"},
		{"Dashboard Development", "Analytics dashboard"},
		{"API Development", "RESTful API endpoints"},
		{"Security Hardening", "Security improvements"},
		{"Performance Optimization", "System performance tuning"},
		{"Mobile Features", "Mobile-specific functionality"},
	}
	
	var ids []string
	for i, epic := range epics {
		projectID := projectIDs[i%len(projectIDs)]
		tags := []string{
			"hub:worca",
			"type:epic",
			"status:in-progress",
			"name:" + epic.name,
			"project:" + projectID,
		}
		content := map[string]interface{}{
			"name":        epic.name,
			"description": epic.desc,
			"priority":    []string{"high", "medium", "low"}[rand.Intn(3)],
		}
		
		id, err := createEntity(token, tags, content)
		if err != nil {
			log.Printf("Failed to create epic %s: %v", epic.name, err)
			continue
		}
		ids = append(ids, id)
		fmt.Printf("  ✓ Created epic: %s\n", epic.name)
	}
	
	return ids
}

func createStories(token string, epicIDs []string, projectIDs []string) []string {
	stories := []struct {
		name string
		desc string
	}{
		{"Login Form Implementation", "Create responsive login form"},
		{"Registration Flow", "User registration workflow"},
		{"Password Reset", "Password reset functionality"},
		{"Two-Factor Authentication", "2FA implementation"},
		{"User Profile Management", "Profile editing features"},
		{"Dashboard Widgets", "Create dashboard components"},
		{"API Endpoints", "REST API implementation"},
		{"Data Visualization", "Charts and graphs"},
		{"Export Functionality", "Data export features"},
		{"Search Implementation", "Full-text search"},
	}
	
	var ids []string
	for i, story := range stories {
		epicID := epicIDs[i%len(epicIDs)]
		projectID := projectIDs[i%len(projectIDs)]
		tags := []string{
			"hub:worca",
			"type:story",
			"status:" + []string{"todo", "in-progress", "done"}[rand.Intn(3)],
			"name:" + story.name,
			"epic:" + epicID,
			"project:" + projectID,
		}
		content := map[string]interface{}{
			"name":         story.name,
			"description":  story.desc,
			"story_points": rand.Intn(8) + 1,
			"priority":     []string{"high", "medium", "low"}[rand.Intn(3)],
		}
		
		id, err := createEntity(token, tags, content)
		if err != nil {
			log.Printf("Failed to create story %s: %v", story.name, err)
			continue
		}
		ids = append(ids, id)
		fmt.Printf("  ✓ Created story: %s\n", story.name)
	}
	
	return ids
}

func createTeamMembers(token string) []string {
	members := []struct {
		name  string
		role  string
		email string
	}{
		{"Alex Johnson", "Full Stack Developer", "alex@company.com"},
		{"Sarah Chen", "UI/UX Designer", "sarah@company.com"},
		{"Mike Rodriguez", "Backend Developer", "mike@company.com"},
		{"Emma Williams", "Product Manager", "emma@company.com"},
		{"David Lee", "DevOps Engineer", "david@company.com"},
		{"Lisa Brown", "QA Engineer", "lisa@company.com"},
		{"Tom Wilson", "Frontend Developer", "tom@company.com"},
		{"Jane Smith", "Data Analyst", "jane@company.com"},
	}
	
	var ids []string
	for _, member := range members {
		tags := []string{
			"hub:worca",
			"type:user",
			"status:active",
			"name:" + member.name,
			"role:" + member.role,
		}
		content := map[string]interface{}{
			"name":         member.name,
			"display_name": member.name,
			"role":         member.role,
			"email":        member.email,
			"department":   "Engineering",
		}
		
		id, err := createEntity(token, tags, content)
		if err != nil {
			log.Printf("Failed to create member %s: %v", member.name, err)
			continue
		}
		ids = append(ids, id)
		fmt.Printf("  ✓ Created team member: %s (%s)\n", member.name, member.role)
	}
	
	return ids
}

func createTasks(token string, storyIDs []string, userIDs []string, projectIDs []string) {
	taskTemplates := []struct {
		title string
		desc  string
		typ   string
	}{
		{"Implement API endpoint", "Create REST endpoint for feature", "backend"},
		{"Design UI mockups", "Create visual designs", "design"},
		{"Write unit tests", "Add test coverage", "testing"},
		{"Update documentation", "Document new features", "documentation"},
		{"Code review", "Review pull request", "review"},
		{"Bug fix", "Fix reported issue", "bug"},
		{"Performance optimization", "Improve response time", "optimization"},
		{"Database migration", "Update schema", "database"},
		{"Security audit", "Check for vulnerabilities", "security"},
		{"Deploy to staging", "Deploy latest changes", "deployment"},
	}
	
	statuses := []string{"todo", "doing", "review", "done"}
	statusWeights := []int{25, 30, 15, 30} // Percentage distribution
	
	// Create 50 tasks with varied distribution
	for i := 0; i < 50; i++ {
		template := taskTemplates[rand.Intn(len(taskTemplates))]
		storyID := storyIDs[rand.Intn(len(storyIDs))]
		userID := userIDs[rand.Intn(len(userIDs))]
		projectID := projectIDs[rand.Intn(len(projectIDs))]
		
		// Weighted status selection
		status := selectWeightedStatus(statuses, statusWeights)
		
		tags := []string{
			"hub:worca",
			"type:task",
			"status:" + status,
			"title:" + fmt.Sprintf("%s #%d", template.title, i+1),
			"story:" + storyID,
			"assignee:" + userID,
			"project:" + projectID,
			"priority:" + []string{"high", "medium", "low"}[rand.Intn(3)],
			"task-type:" + template.typ,
		}
		
		content := map[string]interface{}{
			"title":           fmt.Sprintf("%s #%d", template.title, i+1),
			"description":     template.desc,
			"estimated_hours": rand.Intn(8) + 1,
			"actual_hours":    0,
			"type":            template.typ,
		}
		
		// Add actual hours for completed tasks
		if status == "done" {
			content["actual_hours"] = content["estimated_hours"].(int) + rand.Intn(3) - 1
		}
		
		_, err := createEntity(token, tags, content)
		if err != nil {
			log.Printf("Failed to create task: %v", err)
			continue
		}
		
		if (i+1)%10 == 0 {
			fmt.Printf("  ✓ Created %d tasks...\n", i+1)
		}
	}
	
	fmt.Printf("  ✓ Created 50 tasks total\n")
}

func createHistoricalTasks(token string, storyIDs []string, userIDs []string, projectIDs []string) {
	// Create tasks from the past 30 days for trend analysis
	for daysAgo := 30; daysAgo > 0; daysAgo-- {
		tasksPerDay := rand.Intn(5) + 1
		
		for i := 0; i < tasksPerDay; i++ {
			storyID := storyIDs[rand.Intn(len(storyIDs))]
			userID := userIDs[rand.Intn(len(userIDs))]
			projectID := projectIDs[rand.Intn(len(projectIDs))]
			
			// Historical tasks are mostly done
			status := "done"
			if rand.Float32() < 0.1 { // 10% still in progress
				status = []string{"todo", "doing", "review"}[rand.Intn(3)]
			}
			
			created := time.Now().AddDate(0, 0, -daysAgo)
			
			tags := []string{
				"hub:worca",
				"type:task",
				"status:" + status,
				"title:" + fmt.Sprintf("Historical Task Day-%d-%d", daysAgo, i+1),
				"story:" + storyID,
				"assignee:" + userID,
				"project:" + projectID,
				"priority:" + []string{"high", "medium", "low"}[rand.Intn(3)],
				"created:" + created.Format(time.RFC3339),
			}
			
			content := map[string]interface{}{
				"title":           fmt.Sprintf("Historical Task Day-%d-%d", daysAgo, i+1),
				"description":     "Historical task for metrics",
				"estimated_hours": rand.Intn(8) + 1,
				"actual_hours":    rand.Intn(8) + 1,
				"created_at":      created.Format(time.RFC3339),
				"completed_at":    created.Add(time.Duration(rand.Intn(48)) * time.Hour).Format(time.RFC3339),
			}
			
			_, err := createEntity(token, tags, content)
			if err != nil {
				log.Printf("Failed to create historical task: %v", err)
				continue
			}
		}
	}
	
	fmt.Printf("  ✓ Created historical task data\n")
}

func selectWeightedStatus(statuses []string, weights []int) string {
	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}
	
	r := rand.Intn(totalWeight)
	
	for i, w := range weights {
		r -= w
		if r < 0 {
			return statuses[i]
		}
	}
	
	return statuses[0]
}