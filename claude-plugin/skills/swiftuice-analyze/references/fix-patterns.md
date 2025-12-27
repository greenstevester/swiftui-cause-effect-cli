# SwiftUI Performance Fix Patterns

Complete catalog of fix patterns for SwiftUI performance issues detected by swiftuice.

## Excessive Re-render Fixes

### 1. Implement Equatable on View

**When to use**: View receives data that changes frequently but only some properties affect rendering.

**Effort**: Low | **Impact**: High

```swift
// BEFORE
struct ItemRow: View {
    let item: Item

    var body: some View {
        HStack {
            Text(item.name)
            Spacer()
            Text(item.price, format: .currency(code: "USD"))
        }
    }
}

// AFTER
struct ItemRow: View, Equatable {
    let item: Item

    var body: some View {
        HStack {
            Text(item.name)
            Spacer()
            Text(item.price, format: .currency(code: "USD"))
        }
    }

    static func == (lhs: ItemRow, rhs: ItemRow) -> Bool {
        // Only re-render when these properties change
        lhs.item.id == rhs.item.id &&
        lhs.item.name == rhs.item.name &&
        lhs.item.price == rhs.item.price
    }
}
```

### 2. Extract Subview

**When to use**: Part of a view updates frequently while the rest is static.

**Effort**: Low | **Impact**: Medium

```swift
// BEFORE
struct DashboardView: View {
    @ObservedObject var viewModel: DashboardViewModel

    var body: some View {
        VStack {
            // Static header
            Text("Dashboard")
                .font(.largeTitle)

            // Updates every second
            Text("Updated: \(viewModel.lastUpdate, style: .time)")

            // Static content
            ForEach(viewModel.items) { item in
                ItemRow(item: item)
            }
        }
    }
}

// AFTER
struct DashboardView: View {
    @ObservedObject var viewModel: DashboardViewModel

    var body: some View {
        VStack {
            Text("Dashboard")
                .font(.largeTitle)

            // Extracted to isolate updates
            LastUpdateView(date: viewModel.lastUpdate)

            ForEach(viewModel.items) { item in
                ItemRow(item: item)
            }
        }
    }
}

struct LastUpdateView: View {
    let date: Date

    var body: some View {
        Text("Updated: \(date, style: .time)")
    }
}
```

### 3. Use @Observable (iOS 17+)

**When to use**: Using ObservableObject with many properties but views only need some.

**Effort**: Medium | **Impact**: High

```swift
// BEFORE (iOS 16 and earlier)
class UserViewModel: ObservableObject {
    @Published var name: String = ""
    @Published var email: String = ""
    @Published var avatarURL: URL?
    @Published var lastLoginDate: Date?
    @Published var preferences: UserPreferences = .default
}

struct ProfileHeader: View {
    @ObservedObject var viewModel: UserViewModel

    var body: some View {
        // Re-renders when ANY property changes
        Text(viewModel.name)
    }
}

// AFTER (iOS 17+)
@Observable
class UserViewModel {
    var name: String = ""
    var email: String = ""
    var avatarURL: URL?
    var lastLoginDate: Date?
    var preferences: UserPreferences = .default
}

struct ProfileHeader: View {
    var viewModel: UserViewModel

    var body: some View {
        // Only re-renders when 'name' changes
        Text(viewModel.name)
    }
}
```

## Cascading Update Fixes

### 1. Use Derived State

**When to use**: Multiple views depend on computed values from the same state.

**Effort**: Low | **Impact**: Medium

```swift
// BEFORE
class CartViewModel: ObservableObject {
    @Published var items: [CartItem] = []
    @Published var subtotal: Decimal = 0
    @Published var tax: Decimal = 0
    @Published var total: Decimal = 0

    func updateTotals() {
        subtotal = items.reduce(0) { $0 + $1.price }
        tax = subtotal * 0.08
        total = subtotal + tax
        // Publishing 4 changes causes 4 view updates
    }
}

// AFTER
class CartViewModel: ObservableObject {
    @Published var items: [CartItem] = []

    // Computed - no extra publishes
    var subtotal: Decimal {
        items.reduce(0) { $0 + $1.price }
    }

    var tax: Decimal {
        subtotal * 0.08
    }

    var total: Decimal {
        subtotal + tax
    }
}
```

### 2. Split State Objects

**When to use**: Large ObservableObject with unrelated properties.

**Effort**: Medium | **Impact**: High

```swift
// BEFORE
class AppState: ObservableObject {
    @Published var user: User?
    @Published var cart: [CartItem] = []
    @Published var notifications: [Notification] = []
    @Published var settings: Settings = .default
}

struct CartBadge: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        // Re-renders when user, notifications, or settings change too!
        Text("\(appState.cart.count)")
    }
}

// AFTER
class CartState: ObservableObject {
    @Published var items: [CartItem] = []
}

class NotificationState: ObservableObject {
    @Published var notifications: [Notification] = []
}

struct CartBadge: View {
    @EnvironmentObject var cartState: CartState

    var body: some View {
        // Only re-renders when cart changes
        Text("\(cartState.items.count)")
    }
}
```

### 3. Scope ObservableObject Access

**When to use**: Child views receive parent's ObservableObject but don't need all of it.

**Effort**: Low | **Impact**: Medium

```swift
// BEFORE
struct ProductList: View {
    @ObservedObject var viewModel: StoreViewModel

    var body: some View {
        List(viewModel.products) { product in
            // ProductRow re-renders when ANY viewModel property changes
            ProductRow(viewModel: viewModel, product: product)
        }
    }
}

// AFTER
struct ProductList: View {
    @ObservedObject var viewModel: StoreViewModel

    var body: some View {
        List(viewModel.products) { product in
            // Only pass what's needed
            ProductRow(
                product: product,
                onAddToCart: { viewModel.addToCart(product) }
            )
        }
    }
}

struct ProductRow: View {
    let product: Product
    let onAddToCart: () -> Void

    var body: some View {
        // Only re-renders when this specific product changes
        HStack {
            Text(product.name)
            Button("Add") { onAddToCart() }
        }
    }
}
```

## Timer Cascade Fixes

### 1. Use TimelineView

**When to use**: Content that updates based on time (clocks, countdowns, animations).

**Effort**: Medium | **Impact**: High

```swift
// BEFORE
struct ClockView: View {
    @State private var currentTime = Date()
    let timer = Timer.publish(every: 1, on: .main, in: .common).autoconnect()

    var body: some View {
        VStack {
            Text(currentTime, style: .time)
            // Other content also re-renders every second
            ExpensiveView()
        }
        .onReceive(timer) { time in
            currentTime = time
        }
    }
}

// AFTER
struct ClockView: View {
    var body: some View {
        VStack {
            TimelineView(.periodic(from: .now, by: 1)) { context in
                Text(context.date, style: .time)
            }
            // Other content doesn't re-render
            ExpensiveView()
        }
    }
}
```

### 2. Limit Timer Scope

**When to use**: Timer is at a high level but only one component needs it.

**Effort**: Low | **Impact**: Medium

```swift
// BEFORE
struct ContentView: View {
    @StateObject var viewModel = ContentViewModel()

    var body: some View {
        VStack {
            HeaderView()
            // Timer in viewModel causes all views to re-render
            TimerDisplay(time: viewModel.elapsedTime)
            MainContent()
            FooterView()
        }
    }
}

// AFTER
struct ContentView: View {
    var body: some View {
        VStack {
            HeaderView()
            // Timer isolated in its own view
            StandaloneTimerView()
            MainContent()
            FooterView()
        }
    }
}

struct StandaloneTimerView: View {
    @StateObject private var timer = TimerViewModel()

    var body: some View {
        Text("\(timer.elapsedTime)s")
    }
}
```

## Whole-Object Passing Fixes

### 1. Pass Primitives

**When to use**: View only needs a few properties from a large object.

**Effort**: Low | **Impact**: High

```swift
// BEFORE
struct UserAvatar: View {
    let user: User // User has 20+ properties

    var body: some View {
        AsyncImage(url: user.avatarURL)
            .frame(width: 40, height: 40)
            .clipShape(Circle())
    }
}

// AFTER
struct UserAvatar: View {
    let avatarURL: URL?

    var body: some View {
        AsyncImage(url: avatarURL)
            .frame(width: 40, height: 40)
            .clipShape(Circle())
    }
}

// Usage
UserAvatar(avatarURL: user.avatarURL)
```

### 2. Define Focused Protocol

**When to use**: Multiple views need the same subset of properties.

**Effort**: Medium | **Impact**: Medium

```swift
// BEFORE
struct CommentRow: View {
    let comment: Comment // Has body, author, date, likes, replies, etc.

    var body: some View {
        VStack(alignment: .leading) {
            Text(comment.author.name)
            Text(comment.body)
        }
    }
}

// AFTER
protocol CommentDisplayable {
    var authorName: String { get }
    var bodyText: String { get }
}

extension Comment: CommentDisplayable {
    var authorName: String { author.name }
    var bodyText: String { body }
}

struct CommentRow<T: CommentDisplayable>: View {
    let comment: T

    var body: some View {
        VStack(alignment: .leading) {
            Text(comment.authorName)
            Text(comment.bodyText)
        }
    }
}
```

### 3. Use @Bindable (iOS 17+)

**When to use**: Need to bind to specific properties of an @Observable object.

**Effort**: Low | **Impact**: Medium

```swift
// iOS 17+
@Observable
class FormData {
    var name: String = ""
    var email: String = ""
    var phone: String = ""
    var address: String = ""
}

struct NameField: View {
    @Bindable var formData: FormData

    var body: some View {
        // Only re-renders when name changes
        TextField("Name", text: $formData.name)
    }
}
```

## Performance Verification

After applying fixes, verify improvement:

1. Record a new trace with same user flow
2. Run `swiftuice analyze` again
3. Check:
   - `performance_score` increased
   - `issues_found` decreased
   - Specific view `update_count` reduced
