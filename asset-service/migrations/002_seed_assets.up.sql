INSERT INTO assets (name, type, status, health_score, location) VALUES
('Peloton Bike #1', 'cardio', 'available', 95, 'Zone A'),
('Peloton Bike #2', 'cardio', 'available', 88, 'Zone A'),
('Peloton Tread #1', 'cardio', 'available', 92, 'Zone A'),
('Security Camera Entry', 'security', 'available', 100, 'Entrance'),
('Security Camera Gym Floor', 'security', 'available', 100, 'Main Floor'),
('Smith Machine', 'strength', 'available', 85, 'Zone B'),
('Bench Press #1', 'strength', 'available', 90, 'Zone B'),
('Dumbbell Set 5-50lbs', 'strength', 'available', 98, 'Zone B'),
('Yoga Mat #1', 'yoga', 'available', 75, 'Studio 1'),
('Yoga Mat #2', 'yoga', 'available', 80, 'Studio 1'),
('Kettlebell Set', 'strength', 'available', 95, 'Zone B'),
('Rowing Machine', 'cardio', 'available', 82, 'Zone A'),
('Elliptical #1', 'cardio', 'available', 70, 'Zone A'),
('Leg Press', 'strength', 'available', 88, 'Zone B'),
('Climbing Wall Wall #1', 'climbing', 'available', 99, 'Zone C')
ON CONFLICT DO NOTHING;
