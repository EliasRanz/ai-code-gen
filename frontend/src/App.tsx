import React, { useState, useRef, useEffect } from 'react';
import { generateCode, streamCode, validateCode } from './api';
import { CodeBlock } from './CodeBlock';

export default function App() {
  const [prompt, setPrompt] = useState('');
  const [code, setCode] = useState('');
  const [streaming, setStreaming] = useState(false);
  const [generating, setGenerating] = useState(false);
  const [validating, setValidating] = useState(false);
  const [validation, setValidation] = useState<{ valid: boolean; errors: string[] } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [history, setHistory] = useState<string[]>(() => {
    try {
      const stored = localStorage.getItem('promptHistory');
      return stored ? JSON.parse(stored) : [];
    } catch {
      return [];
    }
  });
  const [theme, setTheme] = useState(() => {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('theme') || 'light';
    }
    return 'light';
  });
  const eventSourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    localStorage.setItem('promptHistory', JSON.stringify(history));
  }, [history]);

  useEffect(() => {
    localStorage.setItem('theme', theme);
  }, [theme]);

  const isDark = theme === 'dark';

  const addToHistory = (newPrompt: string) => {
    setHistory(prev => {
      const filtered = prev.filter(p => p !== newPrompt);
      return [newPrompt, ...filtered].slice(0, 10);
    });
  };

  const handleGenerate = async () => {
    setError(null);
    setValidation(null);
    setCode('');
    setGenerating(true);
    addToHistory(prompt);
    try {
      const res = await generateCode(prompt);
      setCode(res.code);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setGenerating(false);
    }
  };

  const handleStream = () => {
    setError(null);
    setValidation(null);
    setCode('');
    setStreaming(true);
    addToHistory(prompt);
    let fullCode = '';
    eventSourceRef.current = streamCode(prompt, (chunk) => {
      fullCode += chunk;
      setCode(fullCode);
    });
    eventSourceRef.current.onerror = () => {
      setStreaming(false);
      setError('Streaming error or connection closed.');
      eventSourceRef.current?.close();
    };
    eventSourceRef.current.onopen = () => setStreaming(true);
    eventSourceRef.current.onmessage = (event) => {
      if (event.data === '[DONE]') {
        setStreaming(false);
        eventSourceRef.current?.close();
      }
    };
  };

  const handleStopStream = () => {
    eventSourceRef.current?.close();
    setStreaming(false);
  };

  const handleValidate = async () => {
    setError(null);
    setValidation(null);
    setValidating(true);
    try {
      const res = await validateCode(code);
      setValidation({ valid: res.valid, errors: res.errors || [] });
    } catch (e: any) {
      setError(e.message);
    } finally {
      setValidating(false);
    }
  };

  return (
    <div
      style={{
        fontFamily: 'sans-serif',
        maxWidth: 600,
        margin: '2rem auto',
        padding: 24,
        background: isDark ? '#181c1f' : '#fff',
        borderRadius: 16,
        boxShadow: isDark ? '0 2px 16px #0008' : '0 2px 16px #0002',
        minHeight: '90vh',
        display: 'flex',
        flexDirection: 'column',
        gap: 0,
        color: isDark ? '#f1f1f1' : '#222',
        transition: 'background 0.2s, color 0.2s',
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 8 }}>
        <button
          onClick={() => setTheme(isDark ? 'light' : 'dark')}
          style={{
            background: isDark ? '#222' : '#f1f3f6',
            color: isDark ? '#fff' : '#222',
            border: '1px solid #bbb',
            borderRadius: 6,
            padding: '4px 14px',
            fontSize: 14,
            cursor: 'pointer',
            fontWeight: 600,
            marginRight: 0,
            marginBottom: 0,
            transition: 'background 0.2s, color 0.2s',
          }}
          aria-label="Toggle dark/light mode"
        >
          {isDark ? '🌙 Dark' : '☀️ Light'}
        </button>
      </div>
      <h1 style={{ textAlign: 'center', marginBottom: 32, fontWeight: 800, fontSize: 32, letterSpacing: -1 }}>AI UI Generator</h1>
      <textarea
        value={prompt}
        onChange={e => setPrompt(e.target.value)}
        placeholder="Enter your prompt..."
        rows={3}
        style={{
          width: '100%',
          marginBottom: 18,
          fontSize: 16,
          padding: 12,
          borderRadius: 8,
          border: isDark ? '1.5px solid #444' : '1.5px solid #bbb',
          outline: 'none',
          resize: 'vertical',
          boxSizing: 'border-box',
          transition: 'border 0.2s',
          background: isDark ? '#23272b' : '#fff',
          color: isDark ? '#f1f1f1' : '#222',
        }}
        disabled={streaming || generating}
        onFocus={e => (e.currentTarget.style.border = isDark ? '1.5px solid #90caf9' : '1.5px solid #007bff')}
        onBlur={e => (e.currentTarget.style.border = isDark ? '1.5px solid #444' : '1.5px solid #bbb')}
      />
      {history.length > 0 && (
        <div style={{ marginBottom: 10 }}>
          <div style={{ fontSize: 13, color: '#888', marginBottom: 4 }}>Prompt History:</div>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
            {history.map((h, i) => (
              <button
                key={i}
                type="button"
                style={{
                  background: '#f1f3f6',
                  border: '1px solid #e0e0e0',
                  borderRadius: 5,
                  padding: '3px 10px',
                  fontSize: 13,
                  color: '#333',
                  cursor: 'pointer',
                  marginBottom: 2,
                  maxWidth: 180,
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                }}
                onClick={() => setPrompt(h)}
                title={h}
              >
                {h}
              </button>
            ))}
          </div>
        </div>
      )}
      <div
        style={{
          display: 'flex',
          gap: 10,
          marginBottom: 18,
          flexWrap: 'wrap',
          alignItems: 'center',
        }}
      >
        <button
          onClick={handleGenerate}
          disabled={!prompt || streaming || generating}
          style={{
            flex: 1,
            minWidth: 100,
            background: generating ? '#007bff88' : '#007bff',
            color: '#fff',
            border: 'none',
            borderRadius: 6,
            padding: '10px 0',
            fontWeight: 600,
            fontSize: 16,
            cursor: !prompt || streaming || generating ? 'not-allowed' : 'pointer',
            transition: 'background 0.2s',
            boxShadow: generating ? '0 0 0 2px #007bff33' : undefined,
          }}
        >
          {generating ? 'Generating...' : 'Generate'}
        </button>
        <button
          onClick={streaming ? handleStopStream : handleStream}
          disabled={!prompt || generating}
          style={{
            flex: 1,
            minWidth: 100,
            background: streaming ? '#e53935' : '#43a047',
            color: '#fff',
            border: 'none',
            borderRadius: 6,
            padding: '10px 0',
            fontWeight: 600,
            fontSize: 16,
            cursor: !prompt || generating ? 'not-allowed' : 'pointer',
            transition: 'background 0.2s',
            boxShadow: streaming ? '0 0 0 2px #e5393533' : undefined,
          }}
        >
          {streaming ? 'Stop Streaming' : 'Stream'}
        </button>
        <button
          onClick={handleValidate}
          disabled={!code || validating}
          style={{
            flex: 1,
            minWidth: 100,
            background: validating ? '#ffb300' : '#fbc02d',
            color: '#222',
            border: 'none',
            borderRadius: 6,
            padding: '10px 0',
            fontWeight: 600,
            fontSize: 16,
            cursor: !code || validating ? 'not-allowed' : 'pointer',
            transition: 'background 0.2s',
            boxShadow: validating ? '0 0 0 2px #ffb30033' : undefined,
          }}
        >
          {validating ? 'Validating...' : 'Validate'}
        </button>
        {/* Validation status badge */}
        <span
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            marginLeft: 6,
            fontSize: 13,
            fontWeight: 600,
            color: validation == null ? '#888' : validation.valid ? '#43a047' : '#e65100',
            gap: 4,
            minWidth: 70,
          }}
          aria-label="Validation status"
        >
          <span
            style={{
              display: 'inline-block',
              width: 12,
              height: 12,
              borderRadius: '50%',
              background: validation == null ? '#bbb' : validation.valid ? '#43a047' : '#e65100',
              marginRight: 4,
              border: '1.5px solid #fff',
              boxShadow: '0 0 2px #0002',
            }}
          />
          {validation == null ? 'Unchecked' : validation.valid ? 'Valid' : 'Invalid'}
        </span>
      </div>
      {error && <div style={{ color: '#e53935', marginBottom: 14, fontWeight: 500 }}>{error}</div>}
      <div style={{ marginBottom: 18 }}>
        <label style={{ fontWeight: 600, marginBottom: 6, display: 'block', fontSize: 17 }}>
          Generated Code:
          {/* Inline badge for code output */}
          <span
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              marginLeft: 10,
              fontSize: 13,
              fontWeight: 600,
              color: validation == null ? '#888' : validation.valid ? '#43a047' : '#e65100',
              gap: 4,
            }}
            aria-label="Validation status"
          >
            <span
              style={{
                display: 'inline-block',
                width: 10,
                height: 10,
                borderRadius: '50%',
                background: validation == null ? '#bbb' : validation.valid ? '#43a047' : '#e65100',
                marginRight: 3,
                border: '1.5px solid #fff',
                boxShadow: '0 0 2px #0002',
              }}
            />
            {validation == null ? 'Unchecked' : validation.valid ? 'Valid' : 'Invalid'}
          </span>
        </label>
        <div style={{ border: isDark ? '1px solid #333' : '1px solid #eee', borderRadius: 8, background: isDark ? '#23272b' : '#fafbfc', padding: 0, overflow: 'auto' }}>
          <CodeBlock code={code} language="html" />
        </div>
      </div>
      {validation && (
        <div style={{ marginTop: 8, fontWeight: 600, color: validation.valid ? '#43a047' : '#e65100', fontSize: 16 }}>
          {validation.valid ? 'Code is valid!' : 'Validation errors:'}
          {!validation.valid && (
            <ul style={{ margin: '8px 0 0 16px', color: '#e65100', fontWeight: 400, fontSize: 15 }}>
              {validation.errors.map((err, i) => <li key={i}>{err}</li>)}
            </ul>
          )}
        </div>
      )}
      <style>{`
        body { background: ${isDark ? '#101214' : '#f5f5f5'}; }
        @media (max-width: 700px) {
          div[style*='max-width: 600px'] {
            max-width: 98vw !important;
            padding: 8px !important;
          }
        }
        @media (max-width: 500px) {
          h1 {
            font-size: 22px !important;
          }
          textarea {
            font-size: 14px !important;
          }
          button {
            font-size: 14px !important;
            padding: 8px 0 !important;
          }
        }
      `}</style>
    </div>
  );
}
