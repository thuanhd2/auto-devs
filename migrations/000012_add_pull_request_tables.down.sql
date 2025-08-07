-- Drop triggers and functions
DROP TRIGGER IF EXISTS trigger_update_pull_request_checks_updated_at ON pull_request_checks;
DROP FUNCTION IF EXISTS update_pull_request_checks_updated_at();

DROP TRIGGER IF EXISTS trigger_update_pull_request_reviews_updated_at ON pull_request_reviews;
DROP FUNCTION IF EXISTS update_pull_request_reviews_updated_at();

DROP TRIGGER IF EXISTS trigger_update_pull_request_comments_updated_at ON pull_request_comments;
DROP FUNCTION IF EXISTS update_pull_request_comments_updated_at();

DROP TRIGGER IF EXISTS trigger_update_pull_requests_updated_at ON pull_requests;
DROP FUNCTION IF EXISTS update_pull_requests_updated_at();

-- Drop tables in reverse order (child tables first due to foreign keys)
DROP TABLE IF EXISTS pull_request_checks;
DROP TABLE IF EXISTS pull_request_reviews;
DROP TABLE IF EXISTS pull_request_comments;
DROP TABLE IF EXISTS pull_requests;