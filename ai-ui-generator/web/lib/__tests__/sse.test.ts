// Simple validation test for SSE client
import { createMockStreamingClient, SSEOptions } from '../sse'

describe('SSE Client', () => {
  it('should create mock streaming client without errors', () => {
    const options: SSEOptions = {
      onMessage: (data) => console.log('Message:', data),
      onOpen: () => console.log('Connected'),
      onClose: () => console.log('Disconnected'),
      onError: (error) => console.error('Error:', error),
      onStreamEnd: () => console.log('Stream ended')
    }

    const client = createMockStreamingClient(options)
    expect(client).toBeDefined()
    expect(client.isConnected()).toBe(false)
  })
})
