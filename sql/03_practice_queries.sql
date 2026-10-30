-- 1. INNER JOIN: students with their department
SELECT u.full_name, s.student_number, d.name AS department
FROM students s
JOIN users u ON u.id = s.user_id
JOIN departments d ON d.id = s.department_id;

-- 2. LEFT JOIN: all courses, including courses with no enrollments
SELECT c.code, c.title, COUNT(e.id) AS enrollment_count
FROM courses c
LEFT JOIN enrollments e ON e.course_id = c.id
GROUP BY c.id, c.code, c.title
ORDER BY enrollment_count DESC;

-- 3. NOT EXISTS: students who are not enrolled in any course
SELECT s.id, u.full_name
FROM students s
JOIN users u ON u.id = s.user_id
WHERE NOT EXISTS (
    SELECT 1 FROM enrollments e WHERE e.student_id = s.id
);

-- 4. Subquery: students above the global average grade
SELECT DISTINCT s.id, u.full_name
FROM students s
JOIN users u ON u.id = s.user_id
JOIN enrollments e ON e.student_id = s.id
JOIN grades g ON g.enrollment_id = e.id
WHERE g.score > (SELECT AVG(score) FROM grades);

-- 5. GROUP BY and HAVING: departments with at least 5 students
SELECT d.name, COUNT(s.id) AS total_students
FROM departments d
LEFT JOIN students s ON s.department_id = d.id
GROUP BY d.id, d.name
HAVING COUNT(s.id) >= 5;

-- 6. CASE: classify grade results
SELECT enrollment_id, score,
       CASE WHEN score >= 16 THEN 'Very good'
            WHEN score >= 12 THEN 'Good'
            WHEN score >= 10 THEN 'Pass'
            ELSE 'Fail' END AS result
FROM grades;

-- 7. Window function: rank students by weighted average
SELECT student_id, full_name, weighted_average,
       DENSE_RANK() OVER (ORDER BY weighted_average DESC) AS student_rank
FROM student_average
WHERE weighted_average IS NOT NULL;

-- 8. CTE: course performance
WITH course_scores AS (
    SELECT c.id, c.code, c.title, AVG(g.score) AS average_score
    FROM courses c
    JOIN enrollments e ON e.course_id = c.id
    JOIN grades g ON g.enrollment_id = e.id
    GROUP BY c.id, c.code, c.title
)
SELECT *, RANK() OVER (ORDER BY average_score DESC) AS performance_rank
FROM course_scores;

-- 9. Transaction example: enroll a student only if the course exists
BEGIN;
INSERT INTO enrollments(student_id, course_id)
VALUES (1, 1)
ON CONFLICT (student_id, course_id) DO NOTHING;
COMMIT;

-- 10. Audit history
SELECT table_name, operation, row_id, changed_at, old_data, new_data
FROM audit_logs
ORDER BY changed_at DESC;
