package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jessewkun/gocommon/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 测试数据结构
type TestUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type TestResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Status  int         `json:"status"`
}

// 创建测试服务器
func createTestServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test/get":
			handleTestGet(w, r)
		case "/test/post":
			handleTestPost(w, r)
		case "/test/upload":
			handleTestUpload(w, r)
		case "/test/download":
			handleTestDownload(w, r)
		case "/test/timeout":
			handleTestTimeout(w, r)
		case "/test/error":
			handleTestError(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
}

func handleTestGet(w http.ResponseWriter, r *http.Request) {
	user := TestUser{
		ID:    1,
		Name:  "测试用户",
		Email: "test@example.com",
	}

	response := TestResponse{
		Data:    user,
		Message: "获取成功",
		Status:  200,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handleTestPost(w http.ResponseWriter, r *http.Request) {
	var user TestUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 模拟处理逻辑
	user.ID = 100

	response := TestResponse{
		Data:    user,
		Message: "创建成功",
		Status:  201,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func handleTestUpload(w http.ResponseWriter, r *http.Request) {
	// 解析multipart表单
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	response := TestResponse{
		Data: map[string]interface{}{
			"filename":  header.Filename,
			"size":      header.Size,
			"form_data": r.Form,
		},
		Message: "上传成功",
		Status:  200,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handleTestDownload(w http.ResponseWriter, r *http.Request) {
	// 返回一个简单的文件内容
	content := "这是一个测试文件的内容\n包含多行文本\n用于测试下载功能"
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=test.txt")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}

func handleTestTimeout(w http.ResponseWriter, r *http.Request) {
	// 模拟长时间处理
	time.Sleep(5 * time.Second)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("timeout test"))
}

func handleTestError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("服务器内部错误"))
}

// 测试客户端创建
func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		expected bool
	}{
		{
			name: "基本客户端创建",
			option: Option{
				Timeout: 10 * time.Second,
				Headers: map[string]string{
					"User-Agent": "TestClient/1.0",
				},
				IsLog: true,
			},
			expected: true,
		},
		{
			name: "带重试的客户端创建",
			option: Option{
				Timeout:            5 * time.Second,
				Retry:              3,
				RetryWaitTime:      1 * time.Second,
				RetryMaxWaitTime:   5 * time.Second,
				RetryWith5xxStatus: false,
				IsLog:              false,
			},
			expected: true,
		},
		{
			name: "启用5xx重试的客户端创建",
			option: Option{
				Timeout:            5 * time.Second,
				Retry:              3,
				RetryWaitTime:      1 * time.Second,
				RetryMaxWaitTime:   5 * time.Second,
				RetryWith5xxStatus: true,
				IsLog:              false,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.option)
			assert.NotNil(t, client)
			assert.NotNil(t, client.Client)
		})
	}
}

// 测试BuildQuery函数
func TestBuildQuery(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "基本查询参数",
			data: map[string]interface{}{
				"name":  "张三",
				"age":   25,
				"email": "zhangsan@example.com",
			},
			expected: "age=25&email=zhangsan%40example.com&name=%E5%BC%A0%E4%B8%89",
		},
		{
			name:     "空参数",
			data:     map[string]interface{}{},
			expected: "",
		},
		{
			name: "特殊字符参数",
			data: map[string]interface{}{
				"query": "hello world",
				"page":  1,
			},
			expected: "page=1&query=hello+world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildQuery(tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// 测试GET请求
func TestClient_Get(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	client := NewClient(Option{
		Timeout: 10 * time.Second,
		IsLog:   false,
	})

	t.Run("成功GET请求", func(t *testing.T) {
		req := GetRequest{
			URL: server.URL + "/test/get",
			Headers: map[string]string{
				"Accept": "application/json",
			},
		}

		resp, err := client.Get(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, resp.Body)

		// 验证响应内容
		var response TestResponse
		err = json.Unmarshal(resp.Body, &response)
		require.NoError(t, err)
		assert.Equal(t, "获取成功", response.Message)
		assert.Equal(t, 200, response.Status)
	})

	t.Run("带超时的GET请求", func(t *testing.T) {
		req := GetRequest{
			URL:     server.URL + "/test/get",
			Timeout: 5 * time.Second,
		}

		resp, err := client.Get(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("超时GET请求", func(t *testing.T) {
		req := GetRequest{
			URL:     server.URL + "/test/timeout",
			Timeout: 1 * time.Second,
		}

		_, err := client.Get(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	})

	t.Run("错误响应GET请求", func(t *testing.T) {
		req := GetRequest{
			URL: server.URL + "/test/error",
		}

		resp, err := client.Get(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// 测试POST请求
func TestClient_Post(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	client := NewClient(Option{
		Timeout: 10 * time.Second,
		IsLog:   false,
	})

	t.Run("成功POST请求", func(t *testing.T) {
		user := TestUser{
			Name:  "李四",
			Email: "lisi@example.com",
		}

		req := PostRequest{
			URL:     server.URL + "/test/post",
			Payload: user,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}

		resp, err := client.Post(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.NotEmpty(t, resp.Body)

		// 验证响应内容
		var response TestResponse
		err = json.Unmarshal(resp.Body, &response)
		require.NoError(t, err)
		assert.Equal(t, "创建成功", response.Message)
		assert.Equal(t, 201, response.Status)
	})

	t.Run("带超时的POST请求", func(t *testing.T) {
		req := PostRequest{
			URL:     server.URL + "/test/post",
			Payload: map[string]string{"test": "data"},
			Timeout: 5 * time.Second,
		}

		resp, err := client.Post(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

// 测试上传功能
func TestClient_Upload(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	client := NewClient(Option{
		Timeout: 10 * time.Second,
		IsLog:   false,
	})

	t.Run("成功上传文件", func(t *testing.T) {
		fileContent := []byte("这是一个测试文件的内容")
		req := UploadRequest{
			URL:       server.URL + "/test/upload",
			FileBytes: fileContent,
			Param:     "file",
			FileName:  "test.txt",
			Data: map[string]string{
				"description": "测试文件",
				"category":    "test",
			},
			Headers: map[string]string{
				"X-Custom-Header": "test-value",
			},
		}

		resp, err := client.Upload(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, resp.Body)

		// 验证响应内容
		var response TestResponse
		err = json.Unmarshal(resp.Body, &response)
		require.NoError(t, err)
		assert.Equal(t, "上传成功", response.Message)
	})
}

// 测试文件路径上传
func TestClient_UploadWithFilePath(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	client := NewClient(Option{
		Timeout: 10 * time.Second,
		IsLog:   false,
	})

	t.Run("成功上传文件路径", func(t *testing.T) {
		// 创建临时文件
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "test_upload.txt")
		fileContent := []byte("这是通过文件路径上传的测试文件")
		err := os.WriteFile(tempFile, fileContent, 0644)
		require.NoError(t, err)

		req := UploadWithFilePathRequest{
			URL:      server.URL + "/test/upload",
			FilePath: tempFile,
			FileName: "test_upload.txt",
			Param:    "file",
			Data: map[string]string{
				"description": "文件路径上传测试",
			},
		}

		resp, err := client.UploadWithFilePath(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, resp.Body)
	})
}

// 测试下载功能
func TestClient_Download(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	client := NewClient(Option{
		Timeout: 10 * time.Second,
		IsLog:   false,
	})

	t.Run("成功下载文件", func(t *testing.T) {
		tempDir := t.TempDir()
		downloadPath := filepath.Join(tempDir, "downloaded_test.txt")

		req := DownloadRequest{
			URL:      server.URL + "/test/download",
			FilePath: downloadPath,
			Headers: map[string]string{
				"Accept": "text/plain",
			},
		}

		resp, err := client.Download(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// 验证文件是否下载成功
		content, err := os.ReadFile(downloadPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "这是一个测试文件的内容")
	})
}

// 测试真实API请求
func TestClient_RealAPI(t *testing.T) {
	client := NewClient(Option{
		Timeout: 30 * time.Second,
		IsLog:   false,
	})

	t.Run("请求JSONPlaceholder API", func(t *testing.T) {
		req := GetRequest{
			URL: "https://jsonplaceholder.typicode.com/posts/1",
			Headers: map[string]string{
				"Accept": "application/json",
			},
		}

		resp, err := client.Get(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, resp.Body)

		// 验证响应是有效的JSON
		var post map[string]interface{}
		err = json.Unmarshal(resp.Body, &post)
		require.NoError(t, err)
		assert.NotNil(t, post["id"])
		assert.NotNil(t, post["title"])
	})

	t.Run("请求HTTPBin API", func(t *testing.T) {
		testData := map[string]interface{}{
			"name":    "测试用户",
			"email":   "test@example.com",
			"message": "这是一个测试请求",
		}

		req := PostRequest{
			URL:     "https://httpbin.org/post",
			Payload: testData,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "GoCommon-Test/1.0",
			},
		}

		resp, err := client.Post(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, resp.Body)

		// 验证响应包含我们发送的数据
		var response map[string]interface{}
		err = json.Unmarshal(resp.Body, &response)
		require.NoError(t, err)
		assert.NotNil(t, response["json"])
	})

	t.Run("请求HTTPBin获取IP", func(t *testing.T) {
		req := GetRequest{
			URL: "https://httpbin.org/ip",
		}

		resp, err := client.Get(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, resp.Body)

		// 验证响应包含IP信息
		var response map[string]interface{}
		err = json.Unmarshal(resp.Body, &response)
		require.NoError(t, err)
		assert.NotNil(t, response["origin"])
	})
}

// 测试透传参数功能
func TestClient_TransparentParameter(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()

	client := NewClient(Option{
		Timeout: 10 * time.Second,
		IsLog:   false,
	})

	// 设置透传参数
	client.TransparentParameter = []string{"X-User-ID", "X-Trace-ID"}

	t.Run("透传参数测试", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "X-User-ID", "12345")
		ctx = context.WithValue(ctx, "X-Trace-ID", "trace-67890")

		req := GetRequest{
			URL: server.URL + "/test/get",
		}

		resp, err := client.Get(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// 测试重试功能
func TestClient_Retry(t *testing.T) {
	// 创建一个会失败的服务器来测试重试
	retryCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retryCount++
		t.Logf("服务器收到第 %d 次请求", retryCount)
		if retryCount < 3 {
			// 前两次返回500错误
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
		} else {
			// 第三次成功
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	}))
	defer server.Close()

	t.Run("启用5xx重试测试", func(t *testing.T) {
		retryCount = 0 // 重置计数器
		client := NewClient(Option{
			Timeout:            10 * time.Second,
			Retry:              3,
			RetryWaitTime:      100 * time.Millisecond,
			RetryMaxWaitTime:   1 * time.Second,
			RetryWith5xxStatus: true, // 启用5xx重试
			IsLog:              false,
		})

		req := GetRequest{
			URL: server.URL,
		}

		resp, err := client.Get(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "success", string(resp.Body))
		// 验证重试次数，由于启用了5xx重试，应该重试3次
		t.Logf("启用5xx重试 - 实际重试次数: %d", retryCount)
		assert.GreaterOrEqual(t, retryCount, 3)
	})

	t.Run("禁用5xx重试测试", func(t *testing.T) {
		retryCount = 0 // 重置计数器
		client := NewClient(Option{
			Timeout:            10 * time.Second,
			Retry:              3,
			RetryWaitTime:      100 * time.Millisecond,
			RetryMaxWaitTime:   1 * time.Second,
			RetryWith5xxStatus: false, // 禁用5xx重试
			IsLog:              false,
		})

		req := GetRequest{
			URL: server.URL,
		}

		resp, err := client.Get(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode) // 应该返回500错误
		assert.Equal(t, "server error", string(resp.Body))
		// 验证重试次数，由于禁用了5xx重试，应该只请求1次
		t.Logf("禁用5xx重试 - 实际重试次数: %d", retryCount)
		assert.Equal(t, 1, retryCount)
	})
}

// 测试错误处理
func TestClient_ErrorHandling(t *testing.T) {
	client := NewClient(Option{
		Timeout: 5 * time.Second,
		IsLog:   false,
	})

	t.Run("无效URL测试", func(t *testing.T) {
		req := GetRequest{
			URL: "http://invalid-domain-that-does-not-exist-12345.com",
		}

		_, err := client.Get(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("超时测试", func(t *testing.T) {
		req := GetRequest{
			URL:     "https://httpbin.org/delay/10",
			Timeout: 1 * time.Second,
		}

		_, err := client.Get(context.Background(), req)
		assert.Error(t, err)
		// 检查错误信息包含超时相关关键词
		assert.True(t,
			contains(err.Error(), "timeout") ||
				contains(err.Error(), "deadline") ||
				contains(err.Error(), "context"),
			"错误信息应该包含超时相关内容: %s", err.Error())
	})
}

// 辅助函数：检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// 测试日志功能
func TestClient_Logging(t *testing.T) {
	// 初始化logger以避免空指针异常
	err := logger.InitLogger(&logger.Config{
		Path:                 "test.log",
		Closed:               false,
		MaxSize:              100,
		MaxAge:               30,
		MaxBackup:            10,
		TransparentParameter: []string{},
		AlarmLevel:           "warn",
	})
	require.NoError(t, err)

	server := createTestServer(t)
	defer server.Close()

	client := NewClient(Option{
		Timeout: 10 * time.Second,
		IsLog:   true, // 启用日志
	})

	t.Run("日志记录测试", func(t *testing.T) {
		req := GetRequest{
			URL: server.URL + "/test/get",
		}

		resp, err := client.Get(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		// 由于日志是通过logger包记录的，这里主要测试功能是否正常
	})
}

// 基准测试
func BenchmarkClient_Get(b *testing.B) {
	server := createTestServer(nil)
	defer server.Close()

	client := NewClient(Option{
		Timeout: 10 * time.Second,
		IsLog:   false,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := GetRequest{
			URL: server.URL + "/test/get",
		}
		_, err := client.Get(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClient_Post(b *testing.B) {
	server := createTestServer(nil)
	defer server.Close()

	client := NewClient(Option{
		Timeout: 10 * time.Second,
		IsLog:   false,
	})

	testData := map[string]string{
		"name":  "benchmark",
		"email": "bench@test.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := PostRequest{
			URL:     server.URL + "/test/post",
			Payload: testData,
		}
		_, err := client.Post(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
