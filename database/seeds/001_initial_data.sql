-- Initial seed data for Todo App Admin Service
-- This script creates basic system users, categories, and tags

-- System admin user (for system operations)
INSERT INTO users (id, name, email, role, created_at, updated_at, version) VALUES 
(
    '00000000-0000-0000-0000-000000000001',
    'System Admin',
    'admin@todo-app.com',
    'admin',
    NOW(),
    NOW(),
    1
) ON CONFLICT (id) DO NOTHING;

-- Default system user (for auto-created entities)
INSERT INTO users (id, name, email, role, created_at, updated_at, version) VALUES 
(
    '00000000-0000-0000-0000-000000000002',
    'System',
    'system@todo-app.com',
    'admin',
    NOW(),
    NOW(),
    1
) ON CONFLICT (id) DO NOTHING;

-- Test users for development and testing
INSERT INTO users (id, name, email, role, created_at, updated_at, version) VALUES 
(
    '11111111-1111-1111-1111-111111111111',
    'Test Admin',
    'testadmin@example.com',
    'admin',
    NOW(),
    NOW(),
    1
),
(
    '22222222-2222-2222-2222-222222222222',
    'Test User',
    'testuser@example.com',
    'user',
    NOW(),
    NOW(),
    1
),
(
    '33333333-3333-3333-3333-333333333333',
    'John Doe',
    'john@example.com',
    'user',
    NOW(),
    NOW(),
    1
),
(
    '44444444-4444-4444-4444-444444444444',
    'Jane Smith',
    'jane@example.com',
    'user',
    NOW(),
    NOW(),
    1
) ON CONFLICT (id) DO NOTHING;

-- Default categories
INSERT INTO categories (id, name, description, color, creator_id, is_public, created_at, updated_at, version) VALUES 
(
    '10000000-0000-0000-0000-000000000001',
    'Work',
    'Work-related tasks and projects',
    '#3B82F6', -- Blue
    '00000000-0000-0000-0000-000000000001',
    true,
    NOW(),
    NOW(),
    1
),
(
    '10000000-0000-0000-0000-000000000002',
    'Personal',
    'Personal tasks and activities',
    '#10B981', -- Green
    '00000000-0000-0000-0000-000000000001',
    true,
    NOW(),
    NOW(),
    1
),
(
    '10000000-0000-0000-0000-000000000003',
    'Health',
    'Health and wellness related tasks',
    '#F59E0B', -- Yellow
    '00000000-0000-0000-0000-000000000001',
    true,
    NOW(),
    NOW(),
    1
),
(
    '10000000-0000-0000-0000-000000000004',
    'Learning',
    'Educational and skill development tasks',
    '#8B5CF6', -- Purple
    '00000000-0000-0000-0000-000000000001',
    true,
    NOW(),
    NOW(),
    1
),
(
    '10000000-0000-0000-0000-000000000005',
    'Home',
    'Household tasks and maintenance',
    '#EF4444', -- Red
    '00000000-0000-0000-0000-000000000001',
    true,
    NOW(),
    NOW(),
    1
) ON CONFLICT (id) DO NOTHING;

-- Default tags
INSERT INTO tags (id, name, color, creator_id, created_at, updated_at, version) VALUES 
(
    '20000000-0000-0000-0000-000000000001',
    'urgent',
    '#EF4444', -- Red
    '00000000-0000-0000-0000-000000000001',
    NOW(),
    NOW(),
    1
),
(
    '20000000-0000-0000-0000-000000000002',
    'important',
    '#F59E0B', -- Yellow
    '00000000-0000-0000-0000-000000000001',
    NOW(),
    NOW(),
    1
),
(
    '20000000-0000-0000-0000-000000000003',
    'quick',
    '#10B981', -- Green
    '00000000-0000-0000-0000-000000000001',
    NOW(),
    NOW(),
    1
),
(
    '20000000-0000-0000-0000-000000000004',
    'research',
    '#3B82F6', -- Blue
    '00000000-0000-0000-0000-000000000001',
    NOW(),
    NOW(),
    1
),
(
    '20000000-0000-0000-0000-000000000005',
    'meeting',
    '#8B5CF6', -- Purple
    '00000000-0000-0000-0000-000000000001',
    NOW(),
    NOW(),
    1
),
(
    '20000000-0000-0000-0000-000000000006',
    'review',
    '#6B7280', -- Gray
    '00000000-0000-0000-0000-000000000001',
    NOW(),
    NOW(),
    1
),
(
    '20000000-0000-0000-0000-000000000007',
    'bug',
    '#DC2626', -- Dark red
    '00000000-0000-0000-0000-000000000001',
    NOW(),
    NOW(),
    1
),
(
    '20000000-0000-0000-0000-000000000008',
    'feature',
    '#059669', -- Dark green
    '00000000-0000-0000-0000-000000000001',
    NOW(),
    NOW(),
    1
) ON CONFLICT (id) DO NOTHING;

-- Sample tasks for testing
INSERT INTO tasks (id, title, description, assignee_id, status, priority, due_date, created_at, updated_at, version) VALUES 
(
    '30000000-0000-0000-0000-000000000001',
    'Set up development environment',
    'Install and configure all necessary tools for development',
    '33333333-3333-3333-3333-333333333333',
    'COMPLETED',
    'HIGH',
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '1 day',
    1
),
(
    '30000000-0000-0000-0000-000000000002',
    'Review code changes',
    'Review the latest pull request for the admin service',
    '44444444-4444-4444-4444-444444444444',
    'IN_PROGRESS',
    'MEDIUM',
    NOW() + INTERVAL '2 hours',
    NOW() - INTERVAL '1 day',
    NOW(),
    1
),
(
    '30000000-0000-0000-0000-000000000003',
    'Update documentation',
    'Update the API documentation with the latest endpoint changes',
    '22222222-2222-2222-2222-222222222222',
    'OPEN',
    'LOW',
    NOW() + INTERVAL '3 days',
    NOW() - INTERVAL '3 hours',
    NOW() - INTERVAL '3 hours',
    1
) ON CONFLICT (id) DO NOTHING;

-- Associate tasks with categories
INSERT INTO task_categories (task_id, category_id) VALUES 
('30000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001'), -- Work
('30000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000004'), -- Learning
('30000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001'), -- Work
('30000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000001')  -- Work
ON CONFLICT (task_id, category_id) DO NOTHING;

-- Associate tasks with tags
INSERT INTO task_tags (task_id, tag_id) VALUES 
('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000002'), -- important
('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000001'), -- urgent
('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000006'), -- review
('30000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000003')  -- quick
ON CONFLICT (task_id, tag_id) DO NOTHING;

-- Add some task history entries
INSERT INTO task_history (id, task_id, action, actor_id, service_name, timestamp, details) VALUES 
(
    uuid_generate_v4(),
    '30000000-0000-0000-0000-000000000001',
    'created',
    '33333333-3333-3333-3333-333333333333',
    'admin-service',
    NOW() - INTERVAL '2 days',
    '{"status": "OPEN", "priority": "HIGH"}'::jsonb
),
(
    uuid_generate_v4(),
    '30000000-0000-0000-0000-000000000001',
    'status_changed',
    '33333333-3333-3333-3333-333333333333',
    'admin-service',
    NOW() - INTERVAL '1 day',
    '{"from": "OPEN", "to": "COMPLETED"}'::jsonb
),
(
    uuid_generate_v4(),
    '30000000-0000-0000-0000-000000000002',
    'created',
    '11111111-1111-1111-1111-111111111111',
    'admin-service',
    NOW() - INTERVAL '1 day',
    '{"status": "OPEN", "priority": "MEDIUM"}'::jsonb
),
(
    uuid_generate_v4(),
    '30000000-0000-0000-0000-000000000002',
    'status_changed',
    '44444444-4444-4444-4444-444444444444',
    'admin-service',
    NOW() - INTERVAL '2 hours',
    '{"from": "OPEN", "to": "IN_PROGRESS"}'::jsonb
) ON CONFLICT (id) DO NOTHING;