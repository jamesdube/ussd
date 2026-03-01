package ussd

import (
	"log/slog"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jamesdube/ussd/internal/utils"
	"github.com/jamesdube/ussd/pkg/gateway"
	"github.com/jamesdube/ussd/pkg/menu"
	"github.com/jamesdube/ussd/pkg/middleware"
	"github.com/jamesdube/ussd/pkg/router"
	"github.com/jamesdube/ussd/pkg/session"
)

// Mock implementations for testing

type mockGateway struct {
	response interface{}
}

func (m *mockGateway) ToRequest(ctx *fiber.Ctx) (gateway.Request, error) {
	return gateway.Request{
		SessionId: "test-session",
		Message:   "1",
		Msisdn:    "123456789",
	}, nil
}

func (m *mockGateway) Request() gateway.Request {
	return gateway.Request{}
}

func (m *mockGateway) ToResponse(response gateway.Response) interface{} {
	if m.response != nil {
		return m.response
	}
	return "mock response"
}

func (m *mockGateway) Name() string {
	return "mock"
}

type mockMenu struct {
	response       menu.Response
	navigationType menu.NavigationType
}

func (m *mockMenu) OnRequest(c *menu.Context, msg string) menu.Response {
	if m.response.Prompt != "" {
		return m.response
	}
	return menu.Response{
		Prompt:         "mock response",
		NavigationType: menu.Continue,
	}
}

func (m *mockMenu) Process(ctx *menu.Context, msg string) menu.NavigationType {
	if m.navigationType != 0 {
		ctx.NavigationType = m.navigationType
		return m.navigationType
	}
	ctx.NavigationType = menu.Continue
	return menu.Continue
}

// Test helper functions

func createTestFramework() *Framework {
	return &Framework{
		router:             router.NewRouter(),
		middlewareRegistry: middleware.Registry{},
	}
}

func createTestSession(id string) *session.Session {
	if utils.Logger == nil {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		utils.SetLogger(logger)
	}
	sess := session.NewSession(id)
	sess.AddSelection("1")
	return sess
}

func createTestRequestState() *requestState {
	f := createTestFramework()
	ss := createTestSession("test-session")
	c := menu.NewContext("123456789", ss)
	gw := &mockGateway{}

	return &requestState{
		framework: f,
		gateway:   gw,
		request:   gateway.Request{SessionId: "test-session", Message: "1", Msisdn: "123456789"},
		session:   ss,
		context:   c,
	}
}

// --- Navigation Type Tests ---

func TestNavigationTypeHandling(t *testing.T) {
	tests := []struct {
		name           string
		navigationType menu.NavigationType
	}{
		{"Stop navigation should terminate session", menu.Stop},
		{"Replay navigation should go back", menu.Replay},
		{"Continue navigation should proceed normally", menu.Continue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testMenu := &mockMenu{
				navigationType: tt.navigationType,
				response: menu.Response{
					Prompt:         "Test response",
					NavigationType: tt.navigationType,
				},
			}

			sess := createTestSession("test-session")
			c := menu.NewContext("123456789", sess)

			result := testMenu.Process(c, "1")

			if result != tt.navigationType {
				t.Errorf("Expected navigation type %v, got %v", tt.navigationType, result)
			}
			if c.NavigationType != tt.navigationType {
				t.Errorf("Context navigation type not set correctly. Expected %v, got %v",
					tt.navigationType, c.NavigationType)
			}
		})
	}
}

// --- applyPostNavigation Tests ---

func TestApplyPostNavigation(t *testing.T) {
	t.Run("Stop deletes session and deactivates context", func(t *testing.T) {
		s := createTestRequestState()
		s.framework.sessionRepository = session.NewInMemory()
		s.framework.SaveSession(s.session)

		s.applyPostNavigation(menu.Stop)

		if s.context.Active {
			t.Error("Expected context to be inactive after Stop")
		}
	})

	t.Run("Continue saves session", func(t *testing.T) {
		s := createTestRequestState()
		s.framework.sessionRepository = session.NewInMemory()

		s.applyPostNavigation(menu.Continue)

		saved, _ := s.framework.GetSession(s.session.Id)
		if saved == nil {
			t.Error("Expected session to be saved after Continue")
		}
	})

	t.Run("Replay removes two selections and saves", func(t *testing.T) {
		s := createTestRequestState()
		s.framework.sessionRepository = session.NewInMemory()
		s.session.AddSelection("2")
		s.session.AddSelection("3")
		initialCount := len(s.session.GetSelections())

		s.applyPostNavigation(menu.Replay)

		if len(s.session.GetSelections()) != initialCount-2 {
			t.Errorf("Expected %d selections after Replay, got %d",
				initialCount-2, len(s.session.GetSelections()))
		}

		saved, _ := s.framework.GetSession(s.session.Id)
		if saved == nil {
			t.Error("Expected session to be saved after Replay")
		}
	})
}

// --- Session Selection Tests ---

func TestSessionSelectionHandling(t *testing.T) {
	t.Run("Session should add selection in standard navigation", func(t *testing.T) {
		sess := createTestSession("test-session")
		initialSelections := len(sess.GetSelections())

		sess.AddSelection("2")

		if len(sess.GetSelections()) != initialSelections+1 {
			t.Errorf("Expected %d selections, got %d", initialSelections+1, len(sess.GetSelections()))
		}
		if sess.GetSelections()[len(sess.GetSelections())-1] != "2" {
			t.Errorf("Expected last selection to be '2', got '%s'",
				sess.GetSelections()[len(sess.GetSelections())-1])
		}
	})

	t.Run("Session should remove last selection", func(t *testing.T) {
		sess := createTestSession("test-session")
		sess.AddSelection("2")
		sess.AddSelection("3")
		initialCount := len(sess.GetSelections())

		sess.RemoveLastSelection()

		if len(sess.GetSelections()) != initialCount-1 {
			t.Errorf("Expected %d selections after removal, got %d",
				initialCount-1, len(sess.GetSelections()))
		}
	})
}

// --- Menu Context Tests ---

func TestMenuContext(t *testing.T) {
	t.Run("Context should initialize with correct navigation type", func(t *testing.T) {
		sess := createTestSession("test-session")
		c := menu.NewContext("123456789", sess)

		if c.NavigationType != menu.Continue {
			t.Errorf("Expected initial navigation type %v, got %v", menu.Continue, c.NavigationType)
		}
		if c.Msisdn != "123456789" {
			t.Errorf("Expected MSISDN '123456789', got '%s'", c.Msisdn)
		}
		if !c.Active {
			t.Error("Expected context to be active initially")
		}
	})

	t.Run("Context navigation type can be modified", func(t *testing.T) {
		sess := createTestSession("test-session")
		c := menu.NewContext("123456789", sess)

		c.NavigationType = menu.Stop
		if c.NavigationType != menu.Stop {
			t.Errorf("Expected navigation type %v, got %v", menu.Stop, c.NavigationType)
		}

		c.Active = false
		if c.Active {
			t.Error("Expected context to be inactive")
		}
	})
}

// --- buildResponse Tests ---

func TestBuildResponse(t *testing.T) {
	t.Run("Should build response with options", func(t *testing.T) {
		gw := &mockGateway{response: "test response with options"}
		sess := createTestSession("test-session")
		options := []string{"Option 1", "Option 2", "Option 3"}

		result := buildResponse(gw, "Select an option:", options, sess, "123456789", true)

		if result != "test response with options" {
			t.Errorf("Expected 'test response with options', got '%v'", result)
		}
	})

	t.Run("Should build response without options", func(t *testing.T) {
		gw := &mockGateway{response: "test response without options"}
		sess := createTestSession("test-session")

		result := buildResponse(gw, "Final message", nil, sess, "123456789", false)

		if result != "test response without options" {
			t.Errorf("Expected 'test response without options', got '%v'", result)
		}
	})
}

// --- Navigation Flow Integration Tests ---

func TestNavigationFlowIntegration(t *testing.T) {
	t.Run("Complete navigation flow with different types", func(t *testing.T) {
		testCases := []struct {
			processResult   menu.NavigationType
			expectedHandler string
		}{
			{menu.Stop, "handleStop"},
			{menu.Replay, "handleReplay"},
			{menu.Continue, "handleContinue"},
			{menu.Paginated, "handleContinue"},
		}

		for _, tc := range testCases {
			sess := createTestSession("test-session")
			c := menu.NewContext("123456789", sess)
			c.NavigationType = tc.processResult

			if c.NavigationType != tc.processResult {
				t.Errorf("Navigation type not set correctly: expected %v, got %v",
					tc.processResult, c.NavigationType)
			}

			var handlerCalled string
			switch c.NavigationType {
			case menu.Stop:
				handlerCalled = "handleStop"
			case menu.Replay:
				handlerCalled = "handleReplay"
			default:
				handlerCalled = "handleContinue"
			}

			expectedHandler := map[menu.NavigationType]string{
				menu.Stop:      "handleStop",
				menu.Replay:    "handleReplay",
				menu.Continue:  "handleContinue",
				menu.Paginated: "handleContinue",
			}

			if handlerCalled != expectedHandler[tc.processResult] {
				t.Errorf("Wrong handler called for %v: expected %s, got %s",
					tc.processResult, expectedHandler[tc.processResult], handlerCalled)
			}
		}
	})
}

// --- Pagination Tests ---

func TestPaginateOptions(t *testing.T) {
	t.Run("splits options into pages", func(t *testing.T) {
		options := []string{"A", "B", "C", "D", "E"}
		pages := paginateOptions(options, 2)

		if len(pages) != 3 {
			t.Fatalf("Expected 3 pages, got %d", len(pages))
		}
		if len(pages[0]) != 2 {
			t.Errorf("Expected page 0 to have 2 items, got %d", len(pages[0]))
		}
		if len(pages[2]) != 1 {
			t.Errorf("Expected page 2 to have 1 item, got %d", len(pages[2]))
		}
	})

	t.Run("returns single page when perPage is zero", func(t *testing.T) {
		options := []string{"A", "B", "C"}
		pages := paginateOptions(options, 0)

		if len(pages) != 1 {
			t.Fatalf("Expected 1 page, got %d", len(pages))
		}
		if len(pages[0]) != 3 {
			t.Errorf("Expected page to have 3 items, got %d", len(pages[0]))
		}
	})

	t.Run("returns single page when perPage exceeds options", func(t *testing.T) {
		options := []string{"A", "B"}
		pages := paginateOptions(options, 10)

		if len(pages) != 1 {
			t.Fatalf("Expected 1 page, got %d", len(pages))
		}
	})

	t.Run("handles exact page boundary", func(t *testing.T) {
		options := []string{"A", "B", "C", "D"}
		pages := paginateOptions(options, 2)

		if len(pages) != 2 {
			t.Fatalf("Expected 2 pages, got %d", len(pages))
		}
		if len(pages[0]) != 2 || len(pages[1]) != 2 {
			t.Errorf("Expected 2 items per page, got %d and %d", len(pages[0]), len(pages[1]))
		}
	})
}

func TestIsValidPageOption(t *testing.T) {
	s := createTestRequestState()
	s.context.Pages = [][]string{
		{"A", "B", "C"},
		{"D", "E"},
	}

	tests := []struct {
		name      string
		option    int
		pageIndex int
		expected  bool
	}{
		{"valid option on page 0", 1, 0, true},
		{"max valid option on page 0", 3, 0, true},
		{"option too high for page 0", 4, 0, false},
		{"zero option is invalid", 0, 0, false},
		{"valid option on page 1", 2, 1, true},
		{"option too high for page 1", 3, 1, false},
		{"negative page index", 1, -1, false},
		{"page index out of range", 1, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.isValidPageOption(tt.option, tt.pageIndex)
			if result != tt.expected {
				t.Errorf("isValidPageOption(%d, %d) = %v, want %v",
					tt.option, tt.pageIndex, result, tt.expected)
			}
		})
	}
}

func TestGlobalOptionIndex(t *testing.T) {
	s := createTestRequestState()
	s.context.Pages = [][]string{
		{"A", "B", "C"},
		{"D", "E"},
		{"F"},
	}

	tests := []struct {
		name        string
		localOption int
		currentPage int
		expected    int
	}{
		{"first option on first page", 1, 0, 1},
		{"first option on second page", 1, 1, 4},
		{"second option on second page", 2, 1, 5},
		{"first option on third page", 1, 2, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.globalOptionIndex(tt.localOption, tt.currentPage)
			if result != tt.expected {
				t.Errorf("globalOptionIndex(%d, %d) = %d, want %d",
					tt.localOption, tt.currentPage, result, tt.expected)
			}
		})
	}
}

func TestBuildOptions(t *testing.T) {
	t.Run("formats options as numbered list", func(t *testing.T) {
		options := []string{"First", "Second", "Third"}
		result := buildOptions(options)
		expected := "\n1. First\n2. Second\n3. Third"

		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("empty options returns empty string", func(t *testing.T) {
		result := buildOptions([]string{})
		if result != "" {
			t.Errorf("Expected empty string, got %q", result)
		}
	})
}
