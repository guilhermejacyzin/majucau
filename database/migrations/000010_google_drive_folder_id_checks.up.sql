-- PostgreSQL's regex engine rejects bounded repetitions larger than 255.
-- Keep the intended 1..256 identifier bound explicit and validate the
-- allowed alphabet without a repetition count.
ALTER TABLE google_drive_sync_page_checkpoints
  DROP CONSTRAINT google_drive_sync_page_checkpoints_listing_folder_id_check,
  DROP CONSTRAINT google_drive_sync_page_checkpoints_root_folder_id_check,
  ADD CONSTRAINT google_drive_sync_page_checkpoints_listing_folder_id_check
    CHECK (
      listing_folder_id IS NULL OR
      (length(listing_folder_id) BETWEEN 1 AND 256 AND listing_folder_id !~ '[^A-Za-z0-9_-]')
    ),
  ADD CONSTRAINT google_drive_sync_page_checkpoints_root_folder_id_check
    CHECK (
      root_folder_id IS NULL OR
      (length(root_folder_id) BETWEEN 1 AND 256 AND root_folder_id !~ '[^A-Za-z0-9_-]')
    );

ALTER TABLE google_drive_sync_folders
  DROP CONSTRAINT google_drive_sync_folders_root_folder_id_check,
  DROP CONSTRAINT google_drive_sync_folders_folder_id_check,
  DROP CONSTRAINT google_drive_sync_folders_parent_folder_id_check,
  ADD CONSTRAINT google_drive_sync_folders_root_folder_id_check
    CHECK (length(root_folder_id) BETWEEN 1 AND 256 AND root_folder_id !~ '[^A-Za-z0-9_-]'),
  ADD CONSTRAINT google_drive_sync_folders_folder_id_check
    CHECK (length(folder_id) BETWEEN 1 AND 256 AND folder_id !~ '[^A-Za-z0-9_-]'),
  ADD CONSTRAINT google_drive_sync_folders_parent_folder_id_check
    CHECK (
      parent_folder_id IS NULL OR
      (length(parent_folder_id) BETWEEN 1 AND 256 AND parent_folder_id !~ '[^A-Za-z0-9_-]')
    );
