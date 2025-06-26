import React, { useEffect, useRef, useState } from 'react';
import Prism from 'prismjs';
import 'prismjs/themes/prism.css';

export function CodeBlock({ code, language = 'html' }: { code: string; language?: string }) {
  const ref = useRef<HTMLElement>(null);
  const [copied, setCopied] = useState(false);
  useEffect(() => {
    if (ref.current) {
      Prism.highlightElement(ref.current);
    }
  }, [code, language]);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 1200);
    } catch (err) {
      setCopied(false);
    }
  };

  return (
    <div style={{ position: 'relative' }}>
      <button
        onClick={handleCopy}
        style={{
          position: 'absolute',
          top: 8,
          right: 8,
          zIndex: 2,
          background: copied ? '#4caf50' : '#eee',
          color: copied ? '#fff' : '#333',
          border: 'none',
          borderRadius: 4,
          padding: '4px 10px',
          fontSize: 12,
          cursor: 'pointer',
          transition: 'background 0.2s',
        }}
        aria-label="Copy code to clipboard"
      >
        {copied ? 'Copied!' : 'Copy'}
      </button>
      <pre style={{ background: '#f5f5f5', borderRadius: 6, padding: 12, overflowX: 'auto' }}>
        <code ref={ref} className={`language-${language}`}>{code}</code>
      </pre>
    </div>
  );
}
