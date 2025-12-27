// Package suggestions provides fix recommendations for SwiftUI performance issues.
package suggestions

import (
	"github.com/greenstevester/swiftui-cause-effect-cli/internal/issues"
)

// Fix represents a suggested code fix
type Fix struct {
	ID           string   `json:"id"`
	Approach     string   `json:"approach"`
	Description  string   `json:"description"`
	Rationale    string   `json:"rationale"`
	CodeBefore   string   `json:"code_before,omitempty"`
	CodeAfter    string   `json:"code_after,omitempty"`
	Steps        []string `json:"steps"`
	Effort       string   `json:"effort"`       // low, medium, high
	Impact       string   `json:"impact"`       // low, medium, high
	ApplicableTo []string `json:"applicable_to"` // issue types this fix applies to
	SwiftVersion string   `json:"swift_version,omitempty"`
	References   []string `json:"references,omitempty"`
}

// Recommendation is a high-level suggestion for improving performance
type Recommendation struct {
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    int    `json:"priority"` // 1 = highest
}

// GenerateFixes returns applicable fixes for an issue
func GenerateFixes(issue issues.Issue) []Fix {
	var fixes []Fix

	switch issue.Type {
	case issues.IssueExcessiveRerender:
		fixes = append(fixes, getExcessiveRerenderFixes()...)
	case issues.IssueCascadingUpdate:
		fixes = append(fixes, getCascadingUpdateFixes()...)
	case issues.IssueFrequentTrigger:
		fixes = append(fixes, getFrequentTriggerFixes()...)
	case issues.IssueDeepDependencyChain:
		fixes = append(fixes, getDeepChainFixes()...)
	case issues.IssueTimerCascade:
		fixes = append(fixes, getTimerCascadeFixes()...)
	case issues.IssueWholeObjectPassing:
		fixes = append(fixes, getWholeObjectFixes()...)
	}

	return fixes
}

// GenerateRecommendations returns general recommendations based on detected issues
func GenerateRecommendations(detectedIssues []issues.Issue) []Recommendation {
	var recs []Recommendation
	hasIssueType := make(map[issues.IssueType]bool)

	for _, issue := range detectedIssues {
		hasIssueType[issue.Type] = true
	}

	priority := 1

	if hasIssueType[issues.IssueExcessiveRerender] || hasIssueType[issues.IssueCascadingUpdate] {
		recs = append(recs, Recommendation{
			Category:    "Architecture",
			Title:       "Consider using @Observable (iOS 17+)",
			Description: "@Observable provides fine-grained observation - views only update when properties they actually read change, unlike @ObservableObject which triggers on any @Published change.",
			Priority:    priority,
		})
		priority++
	}

	if hasIssueType[issues.IssueWholeObjectPassing] {
		recs = append(recs, Recommendation{
			Category:    "Data Flow",
			Title:       "Pass only required data to child views",
			Description: "Instead of passing entire model objects, extract and pass only the specific properties each view needs. This reduces unnecessary re-renders when unrelated properties change.",
			Priority:    priority,
		})
		priority++
	}

	if hasIssueType[issues.IssueTimerCascade] {
		recs = append(recs, Recommendation{
			Category:    "Animation",
			Title:       "Use TimelineView for time-based updates",
			Description: "TimelineView is optimized for animations and time-based updates. It's more efficient than Timer for UI updates and integrates better with SwiftUI's rendering pipeline.",
			Priority:    priority,
		})
		priority++
	}

	if hasIssueType[issues.IssueDeepDependencyChain] {
		recs = append(recs, Recommendation{
			Category:    "Architecture",
			Title:       "Flatten state dependencies",
			Description: "Deep dependency chains increase latency and make the app harder to debug. Consider using derived state or restructuring to reduce the chain depth.",
			Priority:    priority,
		})
		priority++
	}

	// Always-applicable recommendations
	recs = append(recs, Recommendation{
		Category:    "Debugging",
		Title:       "Use _printChanges() for debugging",
		Description: "Add Self._printChanges() in view body to log exactly why a view is re-rendering. Remove before shipping.",
		Priority:    priority + 10, // Lower priority
	})

	return recs
}

func getExcessiveRerenderFixes() []Fix {
	return []Fix{
		{
			ID:          "equatable-view",
			Approach:    "Implement Equatable on View",
			Description: "Make the view conform to Equatable to control when it re-renders based on meaningful state changes.",
			Rationale:   "SwiftUI can skip re-rendering if it knows the view hasn't meaningfully changed.",
			CodeBefore: `struct ItemRow: View {
    let item: Item

    var body: some View {
        HStack {
            Text(item.name)
            Spacer()
            Text(item.price, format: .currency(code: "USD"))
        }
    }
}`,
			CodeAfter: `struct ItemRow: View, Equatable {
    let item: Item

    static func == (lhs: ItemRow, rhs: ItemRow) -> Bool {
        lhs.item.id == rhs.item.id &&
        lhs.item.name == rhs.item.name &&
        lhs.item.price == rhs.item.price
    }

    var body: some View {
        HStack {
            Text(item.name)
            Spacer()
            Text(item.price, format: .currency(code: "USD"))
        }
    }
}`,
			Steps: []string{
				"Add Equatable conformance to the view struct",
				"Implement == to compare only properties that affect rendering",
				"Wrap usage in EquatableView if needed: EquatableView(content: ItemRow(item: item))",
			},
			Effort:       "low",
			Impact:       "high",
			ApplicableTo: []string{"excessive_rerender"},
		},
		{
			ID:          "extract-subview",
			Approach:    "Extract frequently-updating parts to subviews",
			Description: "Move the frequently-changing content into a separate child view so parent doesn't re-render.",
			Rationale:   "SwiftUI's diffing works at the view level. Smaller views = more granular updates.",
			CodeBefore: `struct ContentView: View {
    @State private var counter = 0
    @State private var items: [Item] = []

    var body: some View {
        VStack {
            Text("Count: \\(counter)")  // Changes frequently
            ForEach(items) { item in    // Expensive, rarely changes
                ItemRow(item: item)
            }
        }
    }
}`,
			CodeAfter: `struct ContentView: View {
    @State private var items: [Item] = []

    var body: some View {
        VStack {
            CounterView()  // Isolated - only this re-renders
            ForEach(items) { item in
                ItemRow(item: item)
            }
        }
    }
}

struct CounterView: View {
    @State private var counter = 0

    var body: some View {
        Text("Count: \\(counter)")
    }
}`,
			Steps: []string{
				"Identify the frequently-changing state",
				"Create a new View struct containing that state",
				"Move the relevant UI code to the new view",
				"Replace the original code with the new subview",
			},
			Effort:       "medium",
			Impact:       "high",
			ApplicableTo: []string{"excessive_rerender", "cascading_update"},
		},
		{
			ID:          "observable-macro",
			Approach:    "Migrate to @Observable (iOS 17+)",
			Description: "Replace @ObservableObject with @Observable for automatic fine-grained observation.",
			Rationale:   "@Observable tracks which properties each view actually reads and only triggers updates for those.",
			CodeBefore: `class UserViewModel: ObservableObject {
    @Published var name: String = ""
    @Published var email: String = ""
    @Published var avatarURL: URL?
}

struct ProfileView: View {
    @ObservedObject var viewModel: UserViewModel
    // Re-renders when ANY property changes
}`,
			CodeAfter: `@Observable
class UserViewModel {
    var name: String = ""
    var email: String = ""
    var avatarURL: URL?
}

struct ProfileView: View {
    var viewModel: UserViewModel
    // Only re-renders when properties used in body change
}`,
			Steps: []string{
				"Replace ObservableObject protocol with @Observable macro",
				"Remove @Published property wrappers",
				"Replace @ObservedObject with plain property or @Bindable",
				"Test that updates still propagate correctly",
			},
			Effort:       "medium",
			Impact:       "high",
			ApplicableTo: []string{"excessive_rerender", "cascading_update", "whole_object_passing"},
			SwiftVersion: "5.9+",
			References:   []string{"https://developer.apple.com/documentation/observation"},
		},
	}
}

func getCascadingUpdateFixes() []Fix {
	return []Fix{
		{
			ID:          "derived-state",
			Approach:    "Use derived/computed state",
			Description: "Instead of storing derived values, compute them from source of truth.",
			Rationale:   "Derived state doesn't need separate updates - it's always consistent with source.",
			CodeBefore: `class ShoppingCart: ObservableObject {
    @Published var items: [CartItem] = []
    @Published var totalPrice: Decimal = 0  // Updated manually
    @Published var itemCount: Int = 0       // Updated manually

    func addItem(_ item: CartItem) {
        items.append(item)
        totalPrice = items.reduce(0) { $0 + $1.price }
        itemCount = items.count
    }
}`,
			CodeAfter: `class ShoppingCart: ObservableObject {
    @Published var items: [CartItem] = []

    var totalPrice: Decimal {
        items.reduce(0) { $0 + $1.price }
    }

    var itemCount: Int {
        items.count
    }

    func addItem(_ item: CartItem) {
        items.append(item)
        // Derived properties update automatically
    }
}`,
			Steps: []string{
				"Identify state that's derived from other state",
				"Convert @Published var to computed var",
				"Remove manual update code",
				"If computation is expensive, consider caching with care",
			},
			Effort:       "low",
			Impact:       "medium",
			ApplicableTo: []string{"cascading_update"},
		},
		{
			ID:          "split-state",
			Approach:    "Split large state objects",
			Description: "Break monolithic state into smaller, focused state objects.",
			Rationale:   "Smaller state objects mean views can subscribe to only what they need.",
			CodeBefore: `class AppState: ObservableObject {
    @Published var user: User?
    @Published var settings: Settings
    @Published var cart: ShoppingCart
    @Published var notifications: [Notification]
    // Every view observing AppState re-renders on any change
}`,
			CodeAfter: `class UserState: ObservableObject {
    @Published var user: User?
}

class SettingsState: ObservableObject {
    @Published var settings: Settings
}

class CartState: ObservableObject {
    @Published var cart: ShoppingCart
}

// Views only observe what they need
struct ProfileView: View {
    @EnvironmentObject var userState: UserState
    // Only re-renders when user changes
}`,
			Steps: []string{
				"Identify logical groupings in your state",
				"Create separate ObservableObject classes for each group",
				"Update views to observe only needed state objects",
				"Consider using @Environment for dependency injection",
			},
			Effort:       "high",
			Impact:       "high",
			ApplicableTo: []string{"cascading_update", "whole_object_passing"},
		},
	}
}

func getFrequentTriggerFixes() []Fix {
	return []Fix{
		{
			ID:          "debounce",
			Approach:    "Debounce rapid updates",
			Description: "Delay processing until updates stop for a short period.",
			Rationale:   "Prevents rapid-fire updates from causing excessive re-renders.",
			CodeBefore: `TextField("Search", text: $searchText)
    .onChange(of: searchText) { newValue in
        performSearch(newValue)  // Fires on every keystroke
    }`,
			CodeAfter: `TextField("Search", text: $searchText)
    .onChange(of: searchText) { newValue in
        searchDebouncer.send(newValue)
    }
    .onReceive(searchDebouncer.debounce(for: .milliseconds(300), scheduler: RunLoop.main)) { value in
        performSearch(value)  // Only fires 300ms after typing stops
    }

// Property:
let searchDebouncer = PassthroughSubject<String, Never>()`,
			Steps: []string{
				"Create a PassthroughSubject for the trigger",
				"Send values to the subject instead of processing directly",
				"Use .debounce() to delay processing",
				"Process values in onReceive after debounce",
			},
			Effort:       "low",
			Impact:       "high",
			ApplicableTo: []string{"frequent_trigger"},
		},
		{
			ID:          "throttle",
			Approach:    "Throttle continuous updates",
			Description: "Limit update frequency to a maximum rate.",
			Rationale:   "Ensures updates happen at most once per interval, even if triggered more often.",
			CodeBefore: `ScrollView {
    // onScroll fires continuously during scroll
}
.onScroll { offset in
    updateHeaderOpacity(for: offset)  // Too frequent
}`,
			CodeAfter: `ScrollView {
    // ...
}
.onScroll { offset in
    scrollThrottler.send(offset)
}
.onReceive(scrollThrottler.throttle(for: .milliseconds(16), scheduler: RunLoop.main, latest: true)) { offset in
    updateHeaderOpacity(for: offset)  // Max 60fps
}`,
			Steps: []string{
				"Create a PassthroughSubject for the event",
				"Send values to the subject on each event",
				"Use .throttle() to limit frequency",
				"Process the latest value at the throttled rate",
			},
			Effort:       "low",
			Impact:       "medium",
			ApplicableTo: []string{"frequent_trigger"},
		},
	}
}

func getDeepChainFixes() []Fix {
	return []Fix{
		{
			ID:          "flatten-hierarchy",
			Approach:    "Flatten the view hierarchy",
			Description: "Reduce nesting levels by combining related views.",
			Rationale:   "Fewer levels means shorter update propagation paths.",
			Steps: []string{
				"Identify deeply nested view hierarchies",
				"Look for wrapper views that only add layout",
				"Combine related views where possible",
				"Use ViewBuilder to compose without nesting",
			},
			Effort:       "medium",
			Impact:       "medium",
			ApplicableTo: []string{"deep_dependency_chain"},
		},
		{
			ID:          "direct-observation",
			Approach:    "Use direct observation instead of passing through",
			Description: "Have child views observe state directly via @EnvironmentObject.",
			Rationale:   "Bypasses intermediate views that would otherwise need to pass data down.",
			CodeBefore: `// State passed through every level
struct GrandparentView: View {
    @StateObject var state = AppState()
    var body: some View {
        ParentView(state: state)
    }
}

struct ParentView: View {
    let state: AppState
    var body: some View {
        ChildView(state: state)  // Just passing through
    }
}`,
			CodeAfter: `// State injected via environment
struct GrandparentView: View {
    @StateObject var state = AppState()
    var body: some View {
        ParentView()
            .environmentObject(state)
    }
}

struct ParentView: View {
    var body: some View {
        ChildView()  // No need to pass state
    }
}

struct ChildView: View {
    @EnvironmentObject var state: AppState
    // Observes directly
}`,
			Steps: []string{
				"Identify state being passed through multiple levels",
				"Inject state using .environmentObject() at appropriate level",
				"Replace parameter passing with @EnvironmentObject",
				"Remove intermediate parameters",
			},
			Effort:       "medium",
			Impact:       "high",
			ApplicableTo: []string{"deep_dependency_chain", "cascading_update"},
		},
	}
}

func getTimerCascadeFixes() []Fix {
	return []Fix{
		{
			ID:          "timeline-view",
			Approach:    "Use TimelineView for animations",
			Description: "TimelineView is optimized for time-based updates and animations.",
			Rationale:   "TimelineView integrates with SwiftUI's rendering pipeline for smooth animations.",
			CodeBefore: `struct ClockView: View {
    @State private var date = Date()
    let timer = Timer.publish(every: 1, on: .main, in: .common).autoconnect()

    var body: some View {
        Text(date, style: .time)
            .onReceive(timer) { date = $0 }
    }
}`,
			CodeAfter: `struct ClockView: View {
    var body: some View {
        TimelineView(.periodic(from: .now, by: 1)) { context in
            Text(context.date, style: .time)
        }
    }
}`,
			Steps: []string{
				"Replace Timer with TimelineView",
				"Choose appropriate schedule (.periodic, .animation, .everyMinute)",
				"Access current time via context.date",
				"Remove @State for time tracking",
			},
			Effort:       "low",
			Impact:       "high",
			ApplicableTo: []string{"timer_cascade"},
			SwiftVersion: "5.5+",
		},
		{
			ID:          "limit-timer-scope",
			Approach:    "Limit timer observation scope",
			Description: "Only the view that needs time should observe the timer.",
			Rationale:   "Prevents timer ticks from cascading to unrelated views.",
			CodeBefore: `struct ParentView: View {
    @State private var time = Date()
    let timer = Timer.publish(every: 1, on: .main, in: .common).autoconnect()

    var body: some View {
        VStack {
            TimeDisplay(time: time)
            ExpensiveListView()  // Re-renders every second!
        }
        .onReceive(timer) { time = $0 }
    }
}`,
			CodeAfter: `struct ParentView: View {
    var body: some View {
        VStack {
            TimeDisplay()  // Timer is isolated here
            ExpensiveListView()  // No longer affected
        }
    }
}

struct TimeDisplay: View {
    @State private var time = Date()
    let timer = Timer.publish(every: 1, on: .main, in: .common).autoconnect()

    var body: some View {
        Text(time, style: .time)
            .onReceive(timer) { time = $0 }
    }
}`,
			Steps: []string{
				"Identify which view actually needs the timer",
				"Move timer and @State to that specific view",
				"Ensure parent views don't hold timer-related state",
			},
			Effort:       "low",
			Impact:       "high",
			ApplicableTo: []string{"timer_cascade"},
		},
	}
}

func getWholeObjectFixes() []Fix {
	return []Fix{
		{
			ID:          "pass-primitives",
			Approach:    "Pass primitive values instead of objects",
			Description: "Extract and pass only the specific properties a view needs.",
			Rationale:   "Primitive properties don't cause re-renders when unrelated object properties change.",
			CodeBefore: `struct UserCard: View {
    let user: User  // Whole object

    var body: some View {
        VStack {
            Text(user.name)
            Text(user.email)
        }
    }
}

// Usage triggers re-render when ANY user property changes
UserCard(user: user)`,
			CodeAfter: `struct UserCard: View {
    let name: String
    let email: String

    var body: some View {
        VStack {
            Text(name)
            Text(email)
        }
    }
}

// Usage only triggers re-render when name or email change
UserCard(name: user.name, email: user.email)`,
			Steps: []string{
				"Identify which properties the view actually uses",
				"Change parameters from object to individual properties",
				"Update call sites to pass specific properties",
				"Consider using a focused protocol if many properties needed",
			},
			Effort:       "low",
			Impact:       "high",
			ApplicableTo: []string{"whole_object_passing", "excessive_rerender"},
		},
		{
			ID:          "focused-protocol",
			Approach:    "Use focused protocols for required data",
			Description: "Define a protocol with only the properties a view needs.",
			Rationale:   "Decouples view from specific model type while documenting requirements.",
			CodeBefore: `struct ItemRow: View {
    let item: Item  // Has 20 properties, view uses 3

    var body: some View {
        HStack {
            Text(item.name)
            Text(item.price, format: .currency(code: "USD"))
            if item.isOnSale { SaleBadge() }
        }
    }
}`,
			CodeAfter: `protocol ItemRowData {
    var name: String { get }
    var price: Decimal { get }
    var isOnSale: Bool { get }
}

extension Item: ItemRowData {}

struct ItemRow<T: ItemRowData>: View {
    let item: T

    var body: some View {
        HStack {
            Text(item.name)
            Text(item.price, format: .currency(code: "USD"))
            if item.isOnSale { SaleBadge() }
        }
    }
}`,
			Steps: []string{
				"Identify properties the view actually reads",
				"Create a protocol with only those properties",
				"Make your model conform to the protocol",
				"Change view to accept the protocol type",
			},
			Effort:       "medium",
			Impact:       "medium",
			ApplicableTo: []string{"whole_object_passing"},
		},
	}
}

// GetAllFixes returns all available fix templates
func GetAllFixes() []Fix {
	var all []Fix
	all = append(all, getExcessiveRerenderFixes()...)
	all = append(all, getCascadingUpdateFixes()...)
	all = append(all, getFrequentTriggerFixes()...)
	all = append(all, getDeepChainFixes()...)
	all = append(all, getTimerCascadeFixes()...)
	all = append(all, getWholeObjectFixes()...)
	return all
}
