CREATE TABLE team_members (
 team_id uuid NOT NULL REFERENCES teams(id), user_id uuid NOT NULL REFERENCES users(id),
 PRIMARY KEY (team_id, user_id)
);
CREATE INDEX team_members_user_idx ON team_members(user_id);
INSERT INTO team_members SELECT team_id, id FROM users WHERE team_id IS NOT NULL ON CONFLICT DO NOTHING;
INSERT INTO team_members SELECT id, leader_id FROM teams WHERE leader_id IS NOT NULL ON CONFLICT DO NOTHING;
ALTER TABLE users ADD COLUMN quote_permissions jsonb DEFAULT '{}'::jsonb;
ALTER TABLE quotes ADD COLUMN team_id uuid REFERENCES teams(id);
CREATE FUNCTION sync_primary_team_membership() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NEW.team_id IS NOT NULL THEN
  INSERT INTO team_members(team_id, user_id) VALUES (NEW.team_id, NEW.id) ON CONFLICT DO NOTHING;
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER users_primary_team_membership AFTER INSERT OR UPDATE OF team_id ON users
 FOR EACH ROW EXECUTE FUNCTION sync_primary_team_membership();
-- A source reference is retained even if the supplier later replaces their lines.
ALTER TABLE quote_line_items ADD COLUMN supplier_quote_item_id uuid;
