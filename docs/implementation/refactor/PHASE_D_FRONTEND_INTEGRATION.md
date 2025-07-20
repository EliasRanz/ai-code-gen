# Phase D: Frontend Integration (CENTRALIZED_AUTH Phase 4)
*Priority: HIGH | Estimated Time: 3-5 days*

## **D.1 NextAuth.js Reconfiguration** ⭐ LEVERAGE CLEAN ARCHITECTURE
**Goal**: Direct frontend integration with centralized auth service

### Implementation Steps:
- [ ] **Update NextAuth configuration** (`/web/lib/auth.ts`):
```typescript
export const authOptions: NextAuthOptions = {
  providers: [
    CredentialsProvider({
      name: "credentials",
      credentials: {
        email: { label: "Email", type: "email" },
        password: { label: "Password", type: "password" }
      },
      async authorize(credentials) {
        if (!credentials?.email || !credentials?.password) {
          return null
        }
        
        try {
          const response = await fetch(`${process.env.AUTH_SERVICE_URL}/api/auth/login`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              email: credentials.email,
              password: credentials.password,
            }),
          })
          
          if (!response.ok) {
            return null
          }
          
          const authResult = await response.json()
          
          return {
            id: authResult.user.id,
            email: authResult.user.email,
            name: authResult.user.username,
            accessToken: authResult.access_token,
            refreshToken: authResult.refresh_token,
          }
        } catch (error) {
          console.error('Auth error:', error)
          return null
        }
      }
    })
  ],
  callbacks: {
    async jwt({ token, user }) {
      if (user) {
        token.accessToken = user.accessToken
        token.refreshToken = user.refreshToken
        token.userId = user.id
      }
      
      // Check if token needs refresh
      if (token.accessToken && isTokenExpired(token.accessToken)) {
        return await refreshAccessToken(token)
      }
      
      return token
    },
    async session({ session, token }) {
      session.accessToken = token.accessToken
      session.userId = token.userId
      session.error = token.error
      return session
    },
  },
  pages: {
    signIn: '/auth/signin',
    error: '/auth/error',
  },
  session: {
    strategy: 'jwt',
    maxAge: 24 * 60 * 60, // 24 hours
  },
}

async function refreshAccessToken(token: JWT): Promise<JWT> {
  try {
    const response = await fetch(`${process.env.AUTH_SERVICE_URL}/api/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        refresh_token: token.refreshToken,
      }),
    })
    
    if (!response.ok) {
      throw new Error('Failed to refresh token')
    }
    
    const refreshedTokens = await response.json()
    
    return {
      ...token,
      accessToken: refreshedTokens.access_token,
      refreshToken: refreshedTokens.refresh_token ?? token.refreshToken,
    }
  } catch (error) {
    return {
      ...token,
      error: 'RefreshAccessTokenError',
    }
  }
}

function isTokenExpired(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return Date.now() >= payload.exp * 1000
  } catch {
    return true
  }
}
```

### Test Requirements:
- [ ] **Auth flow testing**: Complete frontend auth flow
- [ ] **Token refresh testing**: Direct auth service integration
- [ ] **Session management testing**: Auth utilities validation
- [ ] **Error handling testing**: Failed auth scenarios
- [ ] **🏗️ IMPLEMENT MOCK INTEGRATION PATTERN**: Follow ADR-024 mock integration strategy
  - [ ] Generate mocks for frontend service interfaces using `scripts/generate-mocks.sh`
  - [ ] Apply mock patterns for auth service client testing with TypeScript/JavaScript
  - [ ] Use generated mocks for API client and auth flow testing
  - [ ] Reference backend mock integration patterns adapted for frontend testing

### Coding Standards Validation:
- [ ] **File size limits**: Keep all frontend auth files under 300 lines (refactor at 300+, never exceed 500)
- [ ] **Function size limits**: Keep auth functions under 30 lines (refactor at 30+, never exceed 50)
- [ ] **Single responsibility**: Each auth function handles one specific auth operation
- [ ] **Error handling**: Comprehensive error handling for all auth scenarios
- [ ] **Type safety**: Strong typing for all auth-related data structures
- [ ] **Clear separation**: Separate auth logic from UI components
- [ ] **Test coverage**: 90%+ coverage for auth flows and API integration patterns
- [ ] **E2E test coverage**: Complete user journey testing with Cypress or Playwright
- [ ] **Integration test coverage**: API client patterns, command queue functionality, circuit breaker behavior
- [ ] **Performance test coverage**: Auth response times, token refresh performance, API request batching
- [ ] **Accessibility compliance**: WCAG 2.1 AA standards for all auth-related UI components

### Success Criteria:
✅ Frontend authenticates directly with auth service  
✅ Token refresh through auth service  
✅ Clean service boundaries leveraged  
✅ Session management centralized  

### Version Control:
- [ ] **Commit changes**: `git add . && git commit -m "feat: implement NextAuth.js direct auth service integration"`
- [ ] **Validate build**: Ensure all tests pass and frontend builds successfully before committing

---

## **D.2 Frontend API Updates**
**Goal**: Update frontend to use consolidated architecture

### Implementation Steps:
- [ ] **Update API endpoints**: Point to auth service endpoints
- [ ] **Token management**: Direct refresh through auth service
- [ ] **Session utilities**: Update auth helper functions
- [ ] **Error handling**: Consistent error response handling
- [ ] **🏗️ IMPLEMENT API CLIENT INTERFACE PATTERN**: Create consistent API client interface for frontend integrations
  - [ ] Design `APIClient` interface following Java-style pattern for consistent service communication
  - [ ] Implement interface for REST, GraphQL, WebSocket, and future API protocols
  - [ ] Create client factory pattern for dynamic API client selection and configuration
  - [ ] Ensure all API calls follow same interface contract with retry, caching, and error handling
  - [ ] Add client-agnostic authentication, request/response transformation, and offline support
- [ ] **🏗️ IMPLEMENT COMMAND PATTERN**: Add robust request queuing and replay capabilities
  - [ ] Design `APICommand` interface for encapsulating API requests as objects
  - [ ] Implement command queue with priority management for different user tiers
  - [ ] Create command executor with retry logic and circuit breaker integration
  - [ ] Enable request replay for failed operations when connectivity is restored
  - [ ] Add undo/redo functionality for future feature development

### API Client Interface Pattern Design:
```typescript
// Core interface that all API client implementations must implement
interface APIClient {
  // Basic HTTP operations
  get<T>(endpoint: string, params?: QueryParams): Promise<T>
  post<T>(endpoint: string, body?: any): Promise<T>
  put<T>(endpoint: string, body?: any): Promise<T>
  delete<T>(endpoint: string): Promise<T>
  
  // Authentication and headers
  setAuthToken(token: string): void
  setHeaders(headers: Record<string, string>): void
  getHeaders(): Record<string, string>
  
  // Configuration and lifecycle
  configure(config: ClientConfig): void
  setBaseURL(url: string): void
  setRetryPolicy(policy: RetryPolicy): void
  
  // Health and validation
  healthCheck(): Promise<boolean>
  validateConnection(): Promise<boolean>
}

// Factory pattern for API client instantiation
interface ClientFactory {
  createClient(clientType: string, config: ClientConfig): APIClient
  createAuthClient(baseURL: string, authConfig: AuthConfig): APIClient
  createServiceClient(serviceName: string, config: ServiceClientConfig): APIClient
}

// Service-specific client interfaces
interface AuthAPIClient extends APIClient {
  login(credentials: LoginRequest): Promise<AuthResponse>
  logout(): Promise<void>
  refreshToken(): Promise<AuthResponse>
  validateToken(token: string): Promise<UserContext>
}

interface UserAPIClient extends APIClient {
  getUser(userID: string): Promise<User>
  updateUser(user: User): Promise<User>
  getProjects(userID: string): Promise<Project[]>
}

interface AIAPIClient extends APIClient {
  generateCode(req: GenerationRequest): Promise<GenerationResponse>
  streamGeneration(req: GenerationRequest): AsyncIterable<GenerationChunk>
  getGenerationHistory(userID: string): Promise<Generation[]>
}

// Response transformation interface
interface ResponseTransformer {
  transform<T>(response: Response): Promise<T>
  parseError(response: Response): Promise<Error>
}

// Retry and caching interfaces
interface RetryPolicy {
  shouldRetry(attempt: number, error: Error): boolean
  getDelay(attempt: number): number
  getMaxAttempts(): number
}

interface ClientCache {
  get<T>(key: string): T | null
  set<T>(key: string, response: T, ttl: number): void
  delete(key: string): void
  clear(): void
}
```

### Command Pattern Design for API Request Management:
```typescript
// Command pattern for API request queuing, retry, and replay
interface APICommand {
  execute(): Promise<any>
  canRetry(): boolean
  getRetryCount(): number
  incrementRetryCount(): void
  getMaxRetries(): number
  getRequestID(): string
  getPriority(): number
  getTimeout(): number
}

// Concrete command implementations
class AuthLoginCommand implements APICommand {
  private client: APIClient
  private credentials: LoginRequest
  private requestID: string
  private retryCount: number = 0
  private maxRetries: number = 3
  private priority: number = 10 // High priority for auth

  constructor(client: APIClient, credentials: LoginRequest) {
    this.client = client
    this.credentials = credentials
    this.requestID = generateUUID()
  }

  async execute(): Promise<AuthResponse> {
    return this.client.post<AuthResponse>('/api/auth/login', this.credentials)
  }

  canRetry(): boolean {
    return this.retryCount < this.maxRetries
  }

  incrementRetryCount(): void {
    this.retryCount++
  }

  getRequestID(): string {
    return this.requestID
  }

  getPriority(): number {
    return this.priority
  }

  getMaxRetries(): number {
    return this.maxRetries
  }

  getRetryCount(): number {
    return this.retryCount
  }

  getTimeout(): number {
    return 10000 // 10 seconds
  }
}

class CodeGenerationCommand implements APICommand {
  private client: APIClient
  private request: GenerationRequest
  private requestID: string
  private retryCount: number = 0
  private maxRetries: number = 2
  private priority: number = 5 // Normal priority

  constructor(client: APIClient, request: GenerationRequest) {
    this.client = client
    this.request = request
    this.requestID = generateUUID()
    // Higher priority for paying users
    this.priority = request.userTier === 'premium' ? 8 : 5
  }

  async execute(): Promise<GenerationResponse> {
    return this.client.post<GenerationResponse>('/api/ai/generate', this.request)
  }

  canRetry(): boolean {
    return this.retryCount < this.maxRetries
  }

  incrementRetryCount(): void {
    this.retryCount++
  }

  getRequestID(): string {
    return this.requestID
  }

  getPriority(): number {
    return this.priority
  }

  getMaxRetries(): number {
    return this.maxRetries
  }

  getRetryCount(): number {
    return this.retryCount
  }

  getTimeout(): number {
    return 30000 // 30 seconds for AI generation
  }
}

// Command queue with priority and retry management
interface APICommandQueue {
  enqueue(command: APICommand): Promise<void>
  dequeue(): Promise<APICommand | null>
  retryFailed(command: APICommand): Promise<void>
  getQueueSize(): number
  getFailedCommands(): APICommand[]
  clear(): void
}

// Priority queue implementation
class PriorityAPICommandQueue implements APICommandQueue {
  private commands: APICommand[] = []
  private failedQueue: APICommand[] = []
  private maxQueueSize: number = 1000

  async enqueue(command: APICommand): Promise<void> {
    if (this.commands.length >= this.maxQueueSize) {
      throw new Error('Queue is full')
    }

    // Insert command based on priority
    let inserted = false
    for (let i = 0; i < this.commands.length; i++) {
      if (command.getPriority() > this.commands[i].getPriority()) {
        this.commands.splice(i, 0, command)
        inserted = true
        break
      }
    }

    if (!inserted) {
      this.commands.push(command)
    }
  }

  async dequeue(): Promise<APICommand | null> {
    return this.commands.shift() || null
  }

  async retryFailed(command: APICommand): Promise<void> {
    if (command.canRetry()) {
      command.incrementRetryCount()
      await this.enqueue(command)
    } else {
      this.failedQueue.push(command)
    }
  }

  getQueueSize(): number {
    return this.commands.length
  }

  getFailedCommands(): APICommand[] {
    return [...this.failedQueue]
  }

  clear(): void {
    this.commands = []
    this.failedQueue = []
  }
}

// Command executor with circuit breaker integration
class APICommandExecutor {
  private queue: APICommandQueue
  private circuitBreaker: CircuitBreaker
  private metrics: MetricsCollector
  private workers: number = 3
  private isRunning: boolean = false

  constructor(queue: APICommandQueue, circuitBreaker: CircuitBreaker, metrics: MetricsCollector) {
    this.queue = queue
    this.circuitBreaker = circuitBreaker
    this.metrics = metrics
  }

  start(): void {
    if (this.isRunning) return
    
    this.isRunning = true
    for (let i = 0; i < this.workers; i++) {
      this.workerLoop()
    }
  }

  stop(): void {
    this.isRunning = false
  }

  private async workerLoop(): Promise<void> {
    while (this.isRunning) {
      try {
        const command = await this.queue.dequeue()
        if (!command) {
          await new Promise(resolve => setTimeout(resolve, 100)) // Wait for more commands
          continue
        }

        await this.executeWithCircuitBreaker(command)
      } catch (error) {
        console.error('Worker loop error:', error)
      }
    }
  }

  private async executeWithCircuitBreaker(command: APICommand): Promise<void> {
    try {
      const result = await this.circuitBreaker.execute(() => command.execute())
      
      this.metrics.incrementCounter('api_command_success', {
        command_type: command.constructor.name,
      })
    } catch (error) {
      if (command.canRetry()) {
        await this.queue.retryFailed(command)
        this.metrics.incrementCounter('api_command_retry', {
          command_type: command.constructor.name,
        })
      } else {
        this.metrics.incrementCounter('api_command_failed', {
          command_type: command.constructor.name,
        })
      }
    }
  }
}

// Usage in frontend API client
class CommandBasedAPIClient {
  private executor: APICommandExecutor
  private factory: CommandFactory

  constructor(executor: APICommandExecutor, factory: CommandFactory) {
    this.executor = executor
    this.factory = factory
  }

  async login(credentials: LoginRequest): Promise<void> {
    const command = this.factory.createLoginCommand(credentials)
    await this.executor.queue.enqueue(command)
  }

  async generateCode(request: GenerationRequest): Promise<void> {
    const command = this.factory.createGenerationCommand(request)
    await this.executor.queue.enqueue(command)
  }
}
```

### API Client Interface Benefits:
- **Protocol Agnostic**: Switch between REST, GraphQL, WebSocket, or future protocols
- **Consistent API**: All service communication follows same patterns and error handling
- **Retry Logic**: Standardized retry policies with exponential backoff and circuit breakers
- **Caching**: Built-in response caching with TTL and invalidation strategies
- **Authentication**: Automatic token management and refresh across all API calls
- **Testing**: Mock API client implementations for offline and integration testing

### Command Pattern Benefits:
- **Request Queuing**: Queue API requests during network outages for later execution
- **Priority Management**: Higher priority requests (premium users) processed first
- **Retry Logic**: Automatic retry with exponential backoff for failed requests
- **Request Replay**: Replay failed requests when connectivity is restored
- **Undo/Redo**: Potential for request undo functionality in future
- **Metrics**: Detailed tracking of command execution success/failure rates

### Test Requirements:

#### **🧪 E2E Testing Framework Selection**

**Primary Recommendation: Cypress**
```typescript
// cypress/e2e/auth-flow.cy.ts
describe('Authentication Flow', () => {
  it('should complete full auth cycle', () => {
    cy.visit('/auth/signin')
    cy.get('[data-testid="email-input"]').type('test@example.com')
    cy.get('[data-testid="password-input"]').type('password123')
    cy.get('[data-testid="login-button"]').click()
    
    // Verify redirect to dashboard
    cy.url().should('include', '/dashboard')
    cy.get('[data-testid="user-menu"]').should('be.visible')
    
    // Test token refresh
    cy.wait(30000) // Wait for token to near expiry
    cy.get('[data-testid="generate-code"]').click()
    cy.get('[data-testid="generation-result"]').should('be.visible')
  })
})
```

**Alternative Testing Frameworks:**

| Framework | Strengths | Use Cases | Setup Complexity |
|-----------|-----------|-----------|------------------|
| **🥇 Cypress** | Real browser, time travel debugging, excellent DX | Full E2E, component testing | ⭐⭐⭐ Medium |
| **🥈 Playwright** | Multi-browser, fast, parallel execution | Cross-browser E2E, API testing | ⭐⭐ Easy |
| **🥉 TestCafe** | No WebDriver, cross-browser, simple setup | Quick E2E setup, CI/CD integration | ⭐ Very Easy |
| **Puppeteer** | Chrome-focused, programmatic control | Performance testing, PDF generation | ⭐⭐⭐ Medium |
| **WebdriverIO** | Flexible, protocol support, mobile testing | Complex scenarios, mobile apps | ⭐⭐⭐⭐ Complex |

#### **🏗️ E2E TESTING INTERFACE PATTERN**

```typescript
// Create testing framework abstraction for easy switching
interface E2ETestFramework {
  visit(url: string): Promise<void>
  findElement(selector: string): Promise<TestElement>
  type(selector: string, text: string): Promise<void>
  click(selector: string): Promise<void>
  wait(milliseconds: number): Promise<void>
  screenshot(name: string): Promise<void>
  assertVisible(selector: string): Promise<void>
  assertText(selector: string, expectedText: string): Promise<void>
  assertUrl(expectedUrl: string): Promise<void>
  cleanup(): Promise<void>
}

interface TestElement {
  click(): Promise<void>
  type(text: string): Promise<void>
  isVisible(): Promise<boolean>
  getText(): Promise<string>
  getAttribute(name: string): Promise<string>
}

// Cypress Implementation
class CypressTestFramework implements E2ETestFramework {
  async visit(url: string): Promise<void> {
    cy.visit(url)
  }
  
  async findElement(selector: string): Promise<TestElement> {
    return new CypressTestElement(cy.get(selector))
  }
  
  async type(selector: string, text: string): Promise<void> {
    cy.get(selector).type(text)
  }
  
  async click(selector: string): Promise<void> {
    cy.get(selector).click()
  }
  
  async assertVisible(selector: string): Promise<void> {
    cy.get(selector).should('be.visible')
  }
  
  async assertUrl(expectedUrl: string): Promise<void> {
    cy.url().should('include', expectedUrl)
  }
  
  async screenshot(name: string): Promise<void> {
    cy.screenshot(name)
  }
  
  async wait(milliseconds: number): Promise<void> {
    cy.wait(milliseconds)
  }
  
  async cleanup(): Promise<void> {
    cy.clearCookies()
    cy.clearLocalStorage()
  }
}

// Playwright Implementation (alternative)
class PlaywrightTestFramework implements E2ETestFramework {
  private page: Page
  
  constructor(page: Page) {
    this.page = page
  }
  
  async visit(url: string): Promise<void> {
    await this.page.goto(url)
  }
  
  async findElement(selector: string): Promise<TestElement> {
    const element = await this.page.locator(selector)
    return new PlaywrightTestElement(element)
  }
  
  async type(selector: string, text: string): Promise<void> {
    await this.page.fill(selector, text)
  }
  
  async click(selector: string): Promise<void> {
    await this.page.click(selector)
  }
  
  async assertVisible(selector: string): Promise<void> {
    await expect(this.page.locator(selector)).toBeVisible()
  }
  
  async assertUrl(expectedUrl: string): Promise<void> {
    expect(this.page.url()).toContain(expectedUrl)
  }
  
  async screenshot(name: string): Promise<void> {
    await this.page.screenshot({ path: `screenshots/${name}.png` })
  }
  
  async wait(milliseconds: number): Promise<void> {
    await this.page.waitForTimeout(milliseconds)
  }
  
  async cleanup(): Promise<void> {
    await this.page.context().clearCookies()
    await this.page.evaluate(() => localStorage.clear())
  }
}

// Test Framework Factory
class TestFrameworkFactory {
  static createFramework(type: 'cypress' | 'playwright' | 'testcafe', config?: any): E2ETestFramework {
    switch (type) {
      case 'cypress':
        return new CypressTestFramework()
      case 'playwright':
        return new PlaywrightTestFramework(config.page)
      case 'testcafe':
        return new TestCafeTestFramework(config.testController)
      default:
        throw new Error(`Unsupported test framework: ${type}`)
    }
  }
}
```

#### **🧪 Comprehensive Testing Strategy**

**Unit Tests (Jest + React Testing Library)**
```typescript
// components/__tests__/LoginForm.test.tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { LoginForm } from '../LoginForm'
import { MockAPIClient } from '../../../__mocks__/api-client'

describe('LoginForm', () => {
  it('should handle successful login', async () => {
    const mockApiClient = new MockAPIClient()
    mockApiClient.setResponse('/api/auth/login', { success: true, token: 'abc123' })
    
    render(<LoginForm apiClient={mockApiClient} />)
    
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'test@example.com' } })
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'password123' } })
    fireEvent.click(screen.getByRole('button', { name: 'Login' }))
    
    await waitFor(() => {
      expect(screen.getByText('Welcome back!')).toBeInTheDocument()
    })
  })
})
```

**Integration Tests (MSW - Mock Service Worker)**
```typescript
// tests/integration/auth-integration.test.tsx
import { rest } from 'msw'
import { setupServer } from 'msw/node'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { AuthProvider } from '../contexts/AuthContext'
import { App } from '../App'

const server = setupServer(
  rest.post('/api/auth/login', (req, res, ctx) => {
    return res(ctx.json({ success: true, user: { id: '1', email: 'test@example.com' } }))
  }),
  rest.post('/api/auth/refresh', (req, res, ctx) => {
    return res(ctx.json({ access_token: 'new-token', refresh_token: 'new-refresh' }))
  })
)

beforeAll(() => server.listen())
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

describe('Auth Integration', () => {
  it('should integrate with auth service', async () => {
    render(
      <AuthProvider>
        <App />
      </AuthProvider>
    )
    
    // Test full auth flow with mocked service responses
    // ... test implementation
  })
})
```

**E2E Test Categories:**

- [ ] **🔐 Authentication Flow Tests**
  - Login/logout cycles
  - Token refresh scenarios
  - Session timeout handling
  - Multi-tab session management

- [ ] **🚀 API Integration Tests**  
  - Frontend → Auth Service communication
  - Command pattern queue management
  - Priority handling for different user tiers
  - Network failure scenarios and recovery

- [ ] **⚡ Performance Tests**
  - Auth response times (< 2s for login)
  - Token refresh performance (< 1s)
  - Page load times with authentication
  - API request batching efficiency

- [ ] **🔄 Circuit Breaker Tests**
  - Fallback behavior during service outages
  - Recovery when services come back online  
  - User experience during degraded performance
  - Error message display and retry mechanisms

- [ ] **📱 Cross-Browser Compatibility**
  - Chrome, Firefox, Safari, Edge testing
  - Mobile responsive behavior
  - Touch interactions and gestures
  - PWA functionality if applicable

- [ ] **♿ Accessibility Tests**
  - Screen reader compatibility
  - Keyboard navigation
  - ARIA labels and roles
  - Color contrast and visual accessibility

#### **🏗️ Testing Configuration Interface Pattern**

```typescript
// Testing framework configuration abstraction
interface TestConfig {
  framework: TestFramework
  browsers: Browser[]
  baseUrl: string
  apiMocking: MockingStrategy
  reporting: ReportingConfig
  parallel: boolean
  retries: number
  timeout: number
}

type TestFramework = 'cypress' | 'playwright' | 'testcafe' | 'puppeteer'
type Browser = 'chrome' | 'firefox' | 'safari' | 'edge'
type MockingStrategy = 'msw' | 'nock' | 'cypress-intercept' | 'playwright-route'

// Environment-specific test configurations
const testConfigs: Record<string, TestConfig> = {
  development: {
    framework: 'cypress',
    browsers: ['chrome'],
    baseUrl: 'http://localhost:3000',
    apiMocking: 'msw',
    reporting: { format: 'console', screenshots: true },
    parallel: false,
    retries: 1,
    timeout: 10000
  },
  
  ci: {
    framework: 'playwright', // Faster for CI
    browsers: ['chrome', 'firefox'],
    baseUrl: 'http://staging.internal:3000',
    apiMocking: 'playwright-route',
    reporting: { format: 'junit', screenshots: true, videos: true },
    parallel: true,
    retries: 3,
    timeout: 30000
  },
  
  production: {
    framework: 'cypress', // Better debugging for production issues
    browsers: ['chrome', 'firefox', 'safari', 'edge'],
    baseUrl: 'https://app.aicodegen.com',
    apiMocking: 'none', // Real API calls
    reporting: { format: 'dashboard', screenshots: true, videos: true },
    parallel: true,
    retries: 2,
    timeout: 15000
  }
}
```

#### **📊 Test Framework Comparison & Recommendations**

**🥇 Primary: Cypress**
```bash
# Installation
npm install --save-dev cypress @cypress/react @cypress/code-coverage

# Configuration (cypress.config.ts)
export default defineConfig({
  e2e: {
    baseUrl: 'http://localhost:3000',
    supportFile: 'cypress/support/e2e.ts',
    specPattern: 'cypress/e2e/**/*.cy.{js,jsx,ts,tsx}',
    video: true,
    screenshotOnRunFailure: true,
    retries: { runMode: 2, openMode: 0 }
  },
  component: {
    devServer: { framework: 'next', bundler: 'webpack' },
    specPattern: 'src/**/*.cy.{js,jsx,ts,tsx}'
  }
})
```

**Pros:**
- ✅ Excellent developer experience with time travel debugging
- ✅ Real browser testing with visual feedback
- ✅ Component testing capabilities
- ✅ Automatic waiting and retry logic
- ✅ Great documentation and community

**Cons:**
- ❌ Chrome-focused (limited cross-browser)
- ❌ Can be slower than headless alternatives
- ❌ Learning curve for complex scenarios

**🥈 Alternative: Playwright**
```bash
# Installation
npm install --save-dev @playwright/test

# Configuration (playwright.config.ts)
export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure'
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'firefox', use: { ...devices['Desktop Firefox'] } },
    { name: 'webkit', use: { ...devices['Desktop Safari'] } }
  ]
})
```

**Pros:**
- ✅ Multi-browser support (Chrome, Firefox, Safari)
- ✅ Fast execution and parallel testing
- ✅ Built-in API testing capabilities
- ✅ Auto-wait for elements
- ✅ Great CI/CD integration

**Cons:**
- ❌ Less mature ecosystem than Cypress
- ❌ Different API paradigm
- ❌ Limited time travel debugging

**🥉 Quick Setup: TestCafe**
```bash
# Installation
npm install --save-dev testcafe

# Simple test example
import { Selector } from 'testcafe'

fixture('Auth Flow').page('http://localhost:3000/auth/signin')

test('User login', async t => {
  await t
    .typeText('#email', 'test@example.com')
    .typeText('#password', 'password123')
    .click('#login-button')
    .expect(Selector('#dashboard').exists).ok()
})
```

**Pros:**
- ✅ Zero configuration setup
- ✅ Cross-browser without WebDriver
- ✅ Simple API and learning curve
- ✅ Built-in test reports

**Cons:**
- ❌ Limited debugging capabilities
- ❌ Smaller ecosystem
- ❌ Less flexible than Cypress/Playwright

#### **🔧 Recommended Testing Stack**

**Development Environment:**
```yaml
frontend_testing:
  unit_tests:
    framework: "jest"
    testing_library: "@testing-library/react"
    coverage_threshold: 80
  
  integration_tests:
    framework: "jest"
    mocking: "msw"
    api_testing: true
  
  e2e_tests:
    framework: "cypress"
    browsers: ["chrome"]
    component_testing: true
    visual_regression: false
  
  accessibility:
    framework: "@axe-core/react"
    automated_checks: true
```

**CI/CD Environment:**
```yaml
frontend_testing:
  unit_tests:
    framework: "jest"
    coverage_threshold: 90
    parallel: true
  
  e2e_tests:
    framework: "playwright"
    browsers: ["chrome", "firefox"]
    parallel: true
    retries: 3
    artifacts: ["screenshots", "videos", "traces"]
  
  performance_tests:
    framework: "lighthouse-ci"
    metrics: ["performance", "accessibility", "seo"]
    budget_checks: true
  
  visual_regression:
    framework: "percy" # or "chromatic"
    approval_workflow: true
```

### Coding Standards Validation:
- [ ] **File size limits**: Keep all API integration files under 300 lines (refactor at 300+, never exceed 500)
- [ ] **Function size limits**: Keep API functions under 30 lines (refactor at 30+, never exceed 50)
- [ ] **Single responsibility**: Each API function handles one service integration
- [ ] **Error handling**: Robust error handling for all network operations
- [ ] **Retry logic**: Implement appropriate retry mechanisms
- [ ] **Resource cleanup**: Proper cleanup of network resources and timeouts

### Success Criteria:
✅ All frontend auth goes through centralized service  
✅ Consistent auth experience  
✅ Proper error handling and user feedback  
✅ API client interface pattern implemented across all service integrations  
✅ Protocol-agnostic client communication with retry and caching support  
✅ Command pattern implemented for robust request management  
✅ Priority-based request queuing with automatic retry and replay  

### Version Control:
- [ ] **Commit changes**: `git add . && git commit -m "feat: update frontend API integration for centralized auth"`
- [ ] **Validate build**: Ensure all tests pass and frontend builds successfully before committing
