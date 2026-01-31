BEGIN;
INSERT INTO yt_aws_accounts (tenant_id, account_id, account_name, role_arn, external_id, status, created_at)
VALUES (35, '424851482219', 'Test Account 424851482219', 'arn:aws:iam::424851482219:role/YuktiReadOnlyRole', 'yukti-35-RQCqcgGafkc1', 'active', NOW())
RETURNING id;
COMMIT;
