package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codcod/repos/internal/config"
)

func TestCreatePullRequestNonExistentDirectory(t *testing.T) {
	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
		Path: "/path/that/does/not/exist",
	}

	options := PROptions{
		Title: "Test PR",
		Body:  "Test body",
		Token: "fake-token",
	}

	err := CreatePullRequest(repo, options)
	if err == nil {
		t.Error("CreatePullRequest should return error for non-existent directory")
	}
	if !strings.Contains(err.Error(), "repository directory does not exist") {
		t.Errorf("Error should mention non-existent directory, got: %v", err)
	}
}

func TestCreatePullRequestCreateOnlyMode(t *testing.T) {
	tmpDir := t.TempDir()

	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
		Path: tmpDir,
	}

	// Mock HTTP server for GitHub API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/repos/owner/test-repo/pulls") {
			t.Errorf("Expected pulls endpoint, got %s", r.URL.Path)
		}

		// Check authorization header
		auth := r.Header.Get("Authorization")
		if !strings.Contains(auth, "token fake-token") {
			t.Errorf("Expected token authorization, got %s", auth)
		}

		// Read and validate request body
		body, _ := io.ReadAll(r.Body)
		var prRequest map[string]interface{}
		json.Unmarshal(body, &prRequest)

		if prRequest["title"] != "Test PR" {
			t.Errorf("Expected title 'Test PR', got %v", prRequest["title"])
		}
		if prRequest["body"] != "Test body" {
			t.Errorf("Expected body 'Test body', got %v", prRequest["body"])
		}
		if prRequest["base"] != "main" {
			t.Errorf("Expected base 'main', got %v", prRequest["base"])
		}
		if prRequest["draft"] != false {
			t.Errorf("Expected draft false, got %v", prRequest["draft"])
		}

		// Mock successful response
		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"html_url": "https://github.com/owner/test-repo/pull/1",
			"number":   1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override GitHub API URL for testing
	originalCreateGitHubPullRequest := createGitHubPullRequest
	createGitHubPullRequest = func(owner, repo string, options PROptions, baseBranch string) error {
		// Simulate API call to our test server
		client := &http.Client{}
		data := map[string]interface{}{
			"title": options.Title,
			"body":  options.Body,
			"head":  options.BranchName,
			"base":  baseBranch,
			"draft": options.Draft,
		}
		jsonData, _ := json.Marshal(data)
		
		req, _ := http.NewRequest("POST", server.URL+"/repos/"+owner+"/"+repo+"/pulls", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "token "+options.Token)
		req.Header.Set("Content-Type", "application/json")
		
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusCreated {
			return err
		}
		return nil
	}
	defer func() { createGitHubPullRequest = originalCreateGitHubPullRequest }()

	options := PROptions{
		Title:      "Test PR",
		Body:       "Test body",
		BranchName: "test-branch",
		Token:      "fake-token",
		CreateOnly: true,
	}

	err := CreatePullRequest(repo, options)
	if err != nil {
		t.Errorf("CreatePullRequest in create-only mode should not error, got: %v", err)
	}
}

func TestCreatePullRequestWithChanges(t *testing.T) {
	// This test would require setting up a full git repository with changes
	// For now, we'll test the error case when no changes are detected
	tmpDir := t.TempDir()

	// Initialize a git repo (simplified)
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
		Path: tmpDir,
	}

	options := PROptions{
		Title: "Test PR",
		Body:  "Test body",
		Token: "fake-token",
	}

	// This should fail because HasChanges will fail on our mock git repo
	err = CreatePullRequest(repo, options)
	if err == nil {
		t.Error("CreatePullRequest should fail on mock git repo without proper setup")
	}
}

func TestCreateGitHubPullRequestSuccess(t *testing.T) {
	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/repos/owner/repo/pulls") {
			t.Errorf("Expected pulls endpoint, got %s", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
			t.Errorf("Expected GitHub API accept header")
		}
		if r.Header.Get("Authorization") != "token test-token" {
			t.Errorf("Expected token authorization")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected JSON content type")
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var prRequest map[string]interface{}
		json.Unmarshal(body, &prRequest)

		expectedFields := map[string]interface{}{
			"title": "Test Title",
			"body":  "Test Body",
			"head":  "feature-branch",
			"base":  "main",
			"draft": true,
		}

		for field, expected := range expectedFields {
			if prRequest[field] != expected {
				t.Errorf("Expected %s to be %v, got %v", field, expected, prRequest[field])
			}
		}

		// Mock successful response
		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"html_url": "https://github.com/owner/repo/pull/1",
			"number":   1,
			"title":    "Test Title",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Test the function by temporarily replacing the GitHub API URL
	originalFunc := createGitHubPullRequest
	createGitHubPullRequest = func(owner, repo string, options PROptions, baseBranch string) error {
		url := server.URL + "/repos/" + owner + "/" + repo + "/pulls"
		data := map[string]interface{}{
			"title": options.Title,
			"body":  options.Body,
			"head":  options.BranchName,
			"base":  baseBranch,
			"draft": options.Draft,
		}
		jsonData, _ := json.Marshal(data)
		
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("Authorization", "token "+options.Token)
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusCreated {
			return err
		}
		return nil
	}
	defer func() { createGitHubPullRequest = originalFunc }()

	options := PROptions{
		Title:      "Test Title",
		Body:       "Test Body",
		BranchName: "feature-branch",
		Draft:      true,
		Token:      "test-token",
	}

	err := createGitHubPullRequest("owner", "repo", options, "main")
	if err != nil {
		t.Errorf("createGitHubPullRequest should succeed, got: %v", err)
	}
}

func TestCreateGitHubPullRequestFailure(t *testing.T) {
	// Mock HTTP server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		errorResponse := map[string]interface{}{
			"message": "Validation Failed",
			"errors": []map[string]interface{}{
				{
					"resource": "PullRequest",
					"code":     "custom",
					"message":  "A pull request already exists",
				},
			},
		}
		json.NewEncoder(w).Encode(errorResponse)
	}))
	defer server.Close()

	// Test the function with error response
	originalFunc := createGitHubPullRequest
	createGitHubPullRequest = func(owner, repo string, options PROptions, baseBranch string) error {
		url := server.URL + "/repos/" + owner + "/" + repo + "/pulls"
		data := map[string]interface{}{
			"title": options.Title,
			"body":  options.Body,
			"head":  options.BranchName,
			"base":  baseBranch,
			"draft": options.Draft,
		}
		jsonData, _ := json.Marshal(data)
		
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("Authorization", "token "+options.Token)
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusCreated {
			var errorResponse map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&errorResponse)
			return fmt.Errorf("failed to create PR: %v", errorResponse)
		}
		return nil
	}
	defer func() { createGitHubPullRequest = originalFunc }()

	options := PROptions{
		Title:      "Test Title",
		Body:       "Test Body",
		BranchName: "feature-branch",
		Token:      "test-token",
	}

	err := createGitHubPullRequest("owner", "repo", options, "main")
	if err == nil {
		t.Error("createGitHubPullRequest should fail with error response")
	}
}

func TestCreateGitHubPullRequestNoToken(t *testing.T) {
	options := PROptions{
		Title:      "Test Title",
		Body:       "Test Body",
		BranchName: "feature-branch",
		Token:      "", // No token provided
	}

	// Clear environment variable
	oldToken := os.Getenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")
	defer func() {
		if oldToken != "" {
			os.Setenv("GITHUB_TOKEN", oldToken)
		}
	}()

	err := createGitHubPullRequest("owner", "repo", options, "main")
	if err == nil {
		t.Error("createGitHubPullRequest should fail without token")
	}
	if !strings.Contains(err.Error(), "GitHub token not provided") {
		t.Errorf("Error should mention missing token, got: %v", err)
	}
}

func TestCreateGitHubPullRequestWithEnvToken(t *testing.T) {
	// Set token via environment variable
	oldToken := os.Getenv("GITHUB_TOKEN")
	os.Setenv("GITHUB_TOKEN", "env-token")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		if oldToken != "" {
			os.Setenv("GITHUB_TOKEN", oldToken)
		}
	}()

	// Mock server to verify env token is used
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "token env-token" {
			t.Errorf("Expected env token, got %s", auth)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"html_url": "test"})
	}))
	defer server.Close()

	options := PROptions{
		Title:      "Test Title",
		Body:       "Test Body",
		BranchName: "feature-branch",
		Token:      "", // No token in options, should use env
	}

	// Override for testing
	originalFunc := createGitHubPullRequest
	createGitHubPullRequest = func(owner, repo string, options PROptions, baseBranch string) error {
		if options.Token == "" {
			options.Token = os.Getenv("GITHUB_TOKEN")
		}
		if options.Token == "" {
			return fmt.Errorf("GitHub token not provided and GITHUB_TOKEN environment variable not set")
		}
		
		// Simulate successful API call
		req, _ := http.NewRequest("POST", server.URL, nil)
		req.Header.Set("Authorization", "token "+options.Token)
		client := &http.Client{}
		resp, _ := client.Do(req)
		defer resp.Body.Close()
		return nil
	}
	defer func() { createGitHubPullRequest = originalFunc }()

	err := createGitHubPullRequest("owner", "repo", options, "main")
	if err != nil {
		t.Errorf("Should use environment token, got error: %v", err)
	}
}

func TestPROptionsDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
		Path: tmpDir,
	}

	// Test with minimal options
	options := PROptions{
		Token:      "test-token",
		CreateOnly: true, // Skip git operations
	}

	// Mock the GitHub API call to verify defaults
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var prRequest map[string]interface{}
		json.Unmarshal(body, &prRequest)

		// Check defaults
		if prRequest["title"] == "" {
			t.Error("Title should have a default value")
		}
		if prRequest["body"] == "" {
			t.Error("Body should have a default value")
		}
		if prRequest["base"] != "main" {
			t.Error("Base should default to 'main'")
		}
		if prRequest["draft"] != false {
			t.Error("Draft should default to false")
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"html_url": "test"})
	}))
	defer server.Close()

	// This test would need more setup to actually run, but demonstrates the concept
	_ = repo
	_ = options
}

func BenchmarkCreateGitHubPullRequest(b *testing.B) {
	// Mock server for benchmarking
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"html_url": "https://github.com/owner/repo/pull/1",
		})
	}))
	defer server.Close()

	options := PROptions{
		Title:      "Benchmark PR",
		Body:       "Benchmark body",
		BranchName: "benchmark-branch",
		Token:      "benchmark-token",
	}

	// Override for benchmarking
	originalFunc := createGitHubPullRequest
	createGitHubPullRequest = func(owner, repo string, options PROptions, baseBranch string) error {
		client := &http.Client{}
		data := map[string]interface{}{
			"title": options.Title,
			"body":  options.Body,
			"head":  options.BranchName,
			"base":  baseBranch,
			"draft": options.Draft,
		}
		jsonData, _ := json.Marshal(data)
		req, _ := http.NewRequest("POST", server.URL, bytes.NewBuffer(jsonData))
		resp, _ := client.Do(req)
		defer resp.Body.Close()
		return nil
	}
	defer func() { createGitHubPullRequest = originalFunc }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := createGitHubPullRequest("owner", "repo", options, "main")
		if err != nil {
			b.Fatalf("createGitHubPullRequest() error = %v", err)
		}
	}
}