CREATE OR REPLACE VIEW student_average AS
SELECT s.id AS student_id, u.full_name, d.name AS department_name,
       ROUND(SUM(g.score * g.weight) / NULLIF(SUM(g.weight), 0), 2) AS weighted_average,
       COUNT(DISTINCT e.course_id) AS course_count
FROM students s
JOIN users u ON u.id = s.user_id
JOIN departments d ON d.id = s.department_id
LEFT JOIN enrollments e ON e.student_id = s.id AND e.status <> 'withdrawn'
LEFT JOIN grades g ON g.enrollment_id = e.id
GROUP BY s.id, u.full_name, d.name;

CREATE OR REPLACE VIEW course_enrollment_summary AS
SELECT c.id AS course_id, c.code AS course_code, c.title AS course_title,
       COUNT(e.id) FILTER (WHERE e.status = 'active') AS enrollment_count
FROM courses c
LEFT JOIN enrollments e ON e.course_id = c.id
GROUP BY c.id, c.code, c.title;

CREATE OR REPLACE VIEW department_performance AS
SELECT d.id AS department_id, d.name AS department_name,
       COUNT(DISTINCT s.id) AS student_count,
       COUNT(DISTINCT c.id) AS course_count,
       ROUND(AVG(g.score), 2) AS average_grade
FROM departments d
LEFT JOIN students s ON s.department_id = d.id
LEFT JOIN courses c ON c.department_id = d.id
LEFT JOIN enrollments e ON e.student_id = s.id AND e.course_id = c.id
LEFT JOIN grades g ON g.enrollment_id = e.id
GROUP BY d.id, d.name;

CREATE OR REPLACE FUNCTION student_average(p_student_id BIGINT)
RETURNS NUMERIC(5,2)
LANGUAGE SQL
STABLE
AS $$
    SELECT ROUND(SUM(g.score * g.weight) / NULLIF(SUM(g.weight), 0), 2)
    FROM enrollments e
    JOIN grades g ON g.enrollment_id = e.id
    WHERE e.student_id = p_student_id AND e.status <> 'withdrawn';
$$;

CREATE OR REPLACE FUNCTION audit_row_change()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    affected_id BIGINT;
BEGIN
    IF TG_OP = 'DELETE' THEN
        affected_id := OLD.id;
        INSERT INTO audit_logs(table_name, operation, row_id, old_data)
        VALUES (TG_TABLE_NAME, TG_OP, affected_id, to_jsonb(OLD));
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        affected_id := NEW.id;
        INSERT INTO audit_logs(table_name, operation, row_id, old_data, new_data)
        VALUES (TG_TABLE_NAME, TG_OP, affected_id, to_jsonb(OLD), to_jsonb(NEW));
        RETURN NEW;
    ELSE
        affected_id := NEW.id;
        INSERT INTO audit_logs(table_name, operation, row_id, new_data)
        VALUES (TG_TABLE_NAME, TG_OP, affected_id, to_jsonb(NEW));
        RETURN NEW;
    END IF;
END;
$$;

CREATE TRIGGER courses_audit
AFTER INSERT OR UPDATE OR DELETE ON courses
FOR EACH ROW EXECUTE FUNCTION audit_row_change();

CREATE TRIGGER grades_audit
AFTER INSERT OR UPDATE OR DELETE ON grades
FOR EACH ROW EXECUTE FUNCTION audit_row_change();

CREATE OR REPLACE FUNCTION validate_grade_weight()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    existing_weight NUMERIC(6,4);
BEGIN
    SELECT COALESCE(SUM(weight), 0) INTO existing_weight
    FROM grades
    WHERE enrollment_id = NEW.enrollment_id
      AND id <> COALESCE(NEW.id, -1);

    IF existing_weight + NEW.weight > 1.0000 THEN
        RAISE EXCEPTION 'total grade weight cannot exceed 1.0';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER grades_weight_check
BEFORE INSERT OR UPDATE ON grades
FOR EACH ROW EXECUTE FUNCTION validate_grade_weight();
