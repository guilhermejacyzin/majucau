-- The approved future-receipts file importer stores projections under the
-- Nuvem Pago source identity even when its live financial API is unavailable.
-- Keep the integration unavailable for API confirmation while allowing the
-- internal file-import service to obtain its stable connection ID.
INSERT INTO integration_connections (provider, status)
VALUES ('NUVEM_PAGO', 'UNAVAILABLE')
ON CONFLICT (provider) DO NOTHING;
