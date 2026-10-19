-- Give the approved Drive inventory service one stable connection identity.
-- Credentials and folder IDs remain external configuration; this does not
-- claim the Drive account has been authorized or connected.
INSERT INTO integration_connections (provider, status)
VALUES ('GOOGLE_DRIVE', 'NOT_CONFIGURED')
ON CONFLICT (provider) DO NOTHING;
