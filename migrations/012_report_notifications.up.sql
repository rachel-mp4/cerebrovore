CREATE TABLE report_notifications (
  notification_id BIGINT PRIMARY KEY,
  FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
  report_id BIGINT NOT NULL,
  FOREIGN KEY (report_id) REFERENCES reports(id) ON DELETE CASCADE
);
