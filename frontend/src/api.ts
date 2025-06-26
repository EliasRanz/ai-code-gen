export async function generateCode(prompt: string, userId?: string) {
  const res = await fetch('/ai/generate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ prompt, user_id: userId }),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export function streamCode(prompt: string, onChunk: (chunk: string) => void, userId?: string) {
  const url = `/ai/stream/session?prompt=${encodeURIComponent(prompt)}${userId ? `&user_id=${encodeURIComponent(userId)}` : ''}`;
  const eventSource = new EventSource(url);
  eventSource.onmessage = (event) => {
    if (event.data && event.data !== '[DONE]') {
      onChunk(event.data);
    }
  };
  return eventSource;
}

export async function validateCode(code: string) {
  const res = await fetch('/ai/validate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code }),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}
