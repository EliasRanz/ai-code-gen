-- Seed data for AI UI Generator development environment
-- This script populates the database with sample data for testing

-- Insert sample users
INSERT INTO users (id, email, name, avatar_url, roles, is_active, email_verified) VALUES
  ('550e8400-e29b-41d4-a716-446655440001', 'admin@aiuigen.dev', 'Admin User', 'https://avatar.githubusercontent.com/u/1?v=4', ARRAY['admin', 'user'], true, true),
  ('550e8400-e29b-41d4-a716-446655440002', 'developer@aiuigen.dev', 'John Developer', 'https://avatar.githubusercontent.com/u/2?v=4', ARRAY['user'], true, true),
  ('550e8400-e29b-41d4-a716-446655440003', 'designer@aiuigen.dev', 'Jane Designer', 'https://avatar.githubusercontent.com/u/3?v=4', ARRAY['user'], true, true);

-- Insert user settings
INSERT INTO user_settings (user_id, theme, language, notifications, ai_preferences, ui_preferences) VALUES
  ('550e8400-e29b-41d4-a716-446655440001', 'dark', 'en', '{"email": true, "browser": true}', '{"default_model": "gpt-4", "max_tokens": 4000}', '{"sidebar_collapsed": false, "auto_save": true}'),
  ('550e8400-e29b-41d4-a716-446655440002', 'light', 'en', '{"email": true, "browser": false}', '{"default_model": "gpt-3.5-turbo", "max_tokens": 2000}', '{"sidebar_collapsed": true, "auto_save": true}'),
  ('550e8400-e29b-41d4-a716-446655440003', 'auto', 'en', '{"email": false, "browser": true}', '{"default_model": "gpt-4", "max_tokens": 3000}', '{"sidebar_collapsed": false, "auto_save": false}');

-- Insert sample projects
INSERT INTO projects (id, name, description, user_id, status, tags, config, is_public) VALUES
  ('660e8400-e29b-41d4-a716-446655440001', 'E-commerce Dashboard', 'A modern e-commerce admin dashboard with React and TypeScript', '550e8400-e29b-41d4-a716-446655440002', 'active', ARRAY['react', 'typescript', 'dashboard', 'e-commerce'], '{"framework": "react", "typescript": true, "styling": "tailwind", "theme": "dark"}', true),
  ('660e8400-e29b-41d4-a716-446655440002', 'Landing Page Builder', 'Dynamic landing page generator with drag-and-drop components', '550e8400-e29b-41d4-a716-446655440003', 'active', ARRAY['vue', 'landing-page', 'builder'], '{"framework": "vue", "typescript": false, "styling": "css-modules", "theme": "light"}', false),
  ('660e8400-e29b-41d4-a716-446655440003', 'Mobile App UI Kit', 'Complete UI kit for mobile applications', '550e8400-e29b-41d4-a716-446655440002', 'draft', ARRAY['react-native', 'mobile', 'ui-kit'], '{"framework": "react-native", "typescript": true, "platform": "both"}', false),
  ('660e8400-e29b-41d4-a716-446655440004', 'Analytics Platform', 'Data visualization and analytics platform', '550e8400-e29b-41d4-a716-446655440001', 'completed', ARRAY['angular', 'charts', 'analytics'], '{"framework": "angular", "typescript": true, "charts": "d3"}', true);

-- Insert sample chat sessions
INSERT INTO chat_sessions (id, user_id, project_id, title, description, status, context, message_count) VALUES
  ('770e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', 'Dashboard Layout Design', 'Designing the main dashboard layout with sidebar navigation', 'active', '{"current_component": "sidebar", "style_preferences": "modern", "color_scheme": "dark"}', 5),
  ('770e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440002', 'Hero Section Creation', 'Creating an engaging hero section for the landing page', 'active', '{"component_type": "hero", "style": "gradient", "has_cta": true}', 3),
  ('770e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440003', 'Button Components', 'Designing reusable button components for the UI kit', 'completed', '{"component_type": "button", "variants": ["primary", "secondary", "ghost"]}', 8);

-- Insert sample chat messages
INSERT INTO chat_messages (id, chat_session_id, role, type, content, sequence_number, tokens_used, model_used) VALUES
  -- Dashboard Layout Design session
  ('880e8400-e29b-41d4-a716-446655440001', '770e8400-e29b-41d4-a716-446655440001', 'user', 'text', 'I need to create a modern dashboard layout with a collapsible sidebar for an e-commerce admin panel. The design should be clean and use a dark theme.', 1, 0, null),
  ('880e8400-e29b-41d4-a716-446655440002', '770e8400-e29b-41d4-a716-446655440001', 'assistant', 'text', 'I''ll help you create a modern dashboard layout with a collapsible sidebar. Let me design a clean, dark-themed layout that''s perfect for an e-commerce admin panel.', 2, 150, 'gpt-4'),
  ('880e8400-e29b-41d4-a716-446655440003', '770e8400-e29b-41d4-a716-446655440001', 'assistant', 'code', 'import React, { useState } from ''react'';
import { Menu, X, Home, Package, Users, BarChart3 } from ''lucide-react'';

const Dashboard = () => {
  const [sidebarOpen, setSidebarOpen] = useState(true);
  
  return (
    <div className="flex h-screen bg-gray-900 text-white">
      {/* Sidebar */}
      <div className={`${sidebarOpen ? ''w-64'' : ''w-16''} bg-gray-800 transition-all duration-300`}>
        <div className="p-4">
          <div className="flex items-center justify-between">
            {sidebarOpen && <h1 className="text-xl font-bold">Admin Panel</h1>}
            <button onClick={() => setSidebarOpen(!sidebarOpen)}>
              {sidebarOpen ? <X size={24} /> : <Menu size={24} />}
            </button>
          </div>
        </div>
        {/* Navigation items */}
        <nav className="mt-8">
          <div className="px-4 space-y-2">
            <a href="#" className="flex items-center px-4 py-2 rounded hover:bg-gray-700">
              <Home size={20} />
              {sidebarOpen && <span className="ml-3">Dashboard</span>}
            </a>
            <a href="#" className="flex items-center px-4 py-2 rounded hover:bg-gray-700">
              <Package size={20} />
              {sidebarOpen && <span className="ml-3">Products</span>}
            </a>
            <a href="#" className="flex items-center px-4 py-2 rounded hover:bg-gray-700">
              <Users size={20} />
              {sidebarOpen && <span className="ml-3">Customers</span>}
            </a>
            <a href="#" className="flex items-center px-4 py-2 rounded hover:bg-gray-700">
              <BarChart3 size={20} />
              {sidebarOpen && <span className="ml-3">Analytics</span>}
            </a>
          </div>
        </nav>
      </div>
      
      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <header className="bg-gray-800 p-4">
          <h2 className="text-2xl font-semibold">Dashboard Overview</h2>
        </header>
        <main className="flex-1 overflow-y-auto p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {/* Dashboard content goes here */}
          </div>
        </main>
      </div>
    </div>
  );
};

export default Dashboard;', 3, 800, 'gpt-4'),
  ('880e8400-e29b-41d4-a716-446655440004', '770e8400-e29b-41d4-a716-446655440001', 'user', 'text', 'This looks great! Can you add some dashboard cards for key metrics?', 4, 0, null),
  ('880e8400-e29b-41d4-a716-446655440005', '770e8400-e29b-41d4-a716-446655440001', 'assistant', 'code', '// Add these metric cards to the main content area
const MetricCard = ({ title, value, change, icon: Icon }) => (
  <div className="bg-gray-800 p-6 rounded-lg">
    <div className="flex items-center justify-between">
      <div>
        <p className="text-gray-400 text-sm">{title}</p>
        <p className="text-3xl font-bold mt-2">{value}</p>
        <p className={`text-sm mt-2 ${change >= 0 ? ''text-green-400'' : ''text-red-400''}`}>
          {change >= 0 ? ''+'' : ''''}{change}% from last month
        </p>
      </div>
      <Icon size={40} className="text-blue-400" />
    </div>
  </div>
);

// Usage in main content:
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
  <MetricCard title="Total Revenue" value="$124,563" change={12.5} icon={BarChart3} />
  <MetricCard title="Orders" value="1,329" change={-2.1} icon={Package} />
  <MetricCard title="Customers" value="5,421" change={8.3} icon={Users} />
  <MetricCard title="Conversion Rate" value="3.2%" change={0.8} icon={TrendingUp} />
</div>', 5, 400, 'gpt-4'),

  -- Hero Section Creation session
  ('880e8400-e29b-41d4-a716-446655440006', '770e8400-e29b-41d4-a716-446655440002', 'user', 'text', 'I want to create a hero section for a SaaS landing page with a gradient background and call-to-action buttons.', 1, 0, null),
  ('880e8400-e29b-41d4-a716-446655440007', '770e8400-e29b-41d4-a716-446655440002', 'assistant', 'text', 'Perfect! I''ll create a compelling hero section with a beautiful gradient background and prominent call-to-action buttons for your SaaS landing page.', 2, 120, 'gpt-3.5-turbo'),
  ('880e8400-e29b-41d4-a716-446655440008', '770e8400-e29b-41d4-a716-446655440002', 'assistant', 'code', '<section className="relative bg-gradient-to-br from-purple-600 via-blue-600 to-blue-800 min-h-screen flex items-center">
  <div className="absolute inset-0 bg-black opacity-10"></div>
  <div className="relative container mx-auto px-6 text-center text-white">
    <h1 className="text-5xl md:text-7xl font-bold mb-6">
      Build Amazing 
      <span className="text-transparent bg-clip-text bg-gradient-to-r from-yellow-400 to-pink-400">
        {" "}Products{" "}
      </span>
      Faster
    </h1>
    <p className="text-xl md:text-2xl mb-8 max-w-3xl mx-auto text-gray-200">
      Streamline your development workflow with our powerful tools and 
      intuitive interface. Join thousands of developers building the future.
    </p>
    <div className="flex flex-col sm:flex-row gap-4 justify-center">
      <button className="bg-white text-purple-600 px-8 py-4 rounded-full font-semibold text-lg hover:bg-gray-100 transition-colors">
        Start Free Trial
      </button>
      <button className="border-2 border-white text-white px-8 py-4 rounded-full font-semibold text-lg hover:bg-white hover:text-purple-600 transition-colors">
        Watch Demo
      </button>
    </div>
  </div>
</section>', 3, 350, 'gpt-3.5-turbo');

-- Insert sample UI generations
INSERT INTO ui_generations (id, user_id, project_id, chat_session_id, chat_message_id, name, description, prompt, status, component_type, framework, generated_code, metadata) VALUES
  ('990e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', '770e8400-e29b-41d4-a716-446655440001', '880e8400-e29b-41d4-a716-446655440003', 'Dashboard Sidebar', 'Collapsible sidebar navigation for e-commerce dashboard', 'Create a modern dashboard layout with a collapsible sidebar for an e-commerce admin panel. The design should be clean and use a dark theme.', 'completed', 'component', 'react', 'import React, { useState } from ''react''; ...', '{"tokens_used": 800, "model": "gpt-4", "generation_time_ms": 2500}'),
  ('990e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440002', '770e8400-e29b-41d4-a716-446655440002', '880e8400-e29b-41d4-a716-446655440008', 'SaaS Hero Section', 'Hero section with gradient background and CTAs', 'I want to create a hero section for a SaaS landing page with a gradient background and call-to-action buttons.', 'completed', 'component', 'react', '<section className="relative bg-gradient-to-br...', '{"tokens_used": 350, "model": "gpt-3.5-turbo", "generation_time_ms": 1800}');

-- Update chat session message counts and last message times
UPDATE chat_sessions SET 
  message_count = (SELECT COUNT(*) FROM chat_messages WHERE chat_session_id = chat_sessions.id),
  last_message_at = (SELECT MAX(created_at) FROM chat_messages WHERE chat_session_id = chat_sessions.id);

-- Display summary
SELECT 'Database seeded successfully!' as message;
SELECT 
  'Users: ' || COUNT(*) as summary 
FROM users
UNION ALL
SELECT 
  'Projects: ' || COUNT(*) 
FROM projects
UNION ALL
SELECT 
  'Chat Sessions: ' || COUNT(*) 
FROM chat_sessions
UNION ALL
SELECT 
  'Chat Messages: ' || COUNT(*) 
FROM chat_messages
UNION ALL
SELECT 
  'UI Generations: ' || COUNT(*) 
FROM ui_generations;
