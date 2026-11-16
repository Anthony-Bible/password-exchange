# Runbook: prod data MySQL → Postgres (#569)

Moves password-exchange's prod data from the Cloud SQL MySQL instance `prod`
(shared with other legacy apps) to the `passwordexchange` and `slack_oauth`
databases on the Cloud SQL Postgres instance `sre-bible`. One Secret switch
(`password-exchange-db-prod`) moves the database service, the prod reminder
CronJob and the slackbot together. Part of epic #562; #570 retires MySQL afterwards.

Tooling lives in [`deploy/migration/`](../../deploy/migration/):

| file | what it does |
|---|---|
| `copy.sh` | stages `preflight`, `prepare`, `load`, `verify`, `reset`; the rehearsal and the cutover run the same script |
| `loadfiles.py` | renders the three pgloader load files (0600, passwords URL-encoded) |
| `sequences.py` | sets each identity sequence from MySQL `AUTO_INCREMENT` |
| `verify.py` | row-by-row diff of every copied table, plus a sequence check; prints counts only |
| `slack_tables.py`, `slack_lookup.py` | create the slack-sdk tables, and run `find_bot`/`find_installation` through the slackbot's own store |

Tests: `pytest deploy/migration/tests` (they also run in CI).

## Decisions

| | |
|---|---|
| Cutover date | on or after **2026-11-18**, in a window the operator picks |
| Rows | copy **everything**, including dev-era leftovers (dev and prod shared this DB before #568) |
| Operator path | laptop + podman: two `cloud-sql-proxy` containers and pgloader in one pod |
| Rollback window | **24 h** after cutover; after that, fix forward only |
| MySQL write grants | revoked **after** the 24 h window, once nothing else is shown to use these users |
| Slack check | operator runs `slack_lookup.py`, then a human runs `/password` in Slack |

## What moves

| MySQL db.table | → Postgres db.table | rows at pre-flight (2026-11-16) |
|---|---|---|
| `passwordExchange.messages` | `passwordexchange.messages` | 67 |
| `passwordExchange.email_reminders` | `passwordexchange.email_reminders` | 0 |
| `passwordExchange.file_upload_sessions` | `passwordexchange.file_upload_sessions` | 56 |
| `oauth.slack_installations` | `slack_oauth.slack_installations` | 11 (2 teams) |
| `oauth.slack_bots` | `slack_oauth.slack_bots` | 11 |

Not copied:
- `schema_migrations`: created by `app database migrate up`, which `prepare` gates on matching MySQL's version.
- `slack_oauth_states`: short-lived OAuth `state` values; any in-flight install just restarts.

The target schema is created **before** the load by the app's own migrations and the
slackbot's own `create_all`. pgloader runs data-only and must not change it. `load`
diffs `pg_dump --schema-only` before and after, and stops if they differ.

## Pre-flight findings (read-only, 2026-11-16)

- MySQL 8.0.41. `@@global.time_zone` = SYSTEM and `@@system_time_zone` = UTC, so
  `TIMESTAMP`/`DATETIME` values are read as UTC. `preflight` re-checks this, because
  the load files no longer set a session time zone.
- `messages`, `email_reminders` and the Slack tables are latin1; `file_upload_sessions` is
  utf8mb4. There are no non-ASCII bytes in any text column, so the conversion is lossless.
- No zero dates, NUL bytes, invalid JSON, orphaned reminders or expired messages. The
  longest TTL is about 87 days, which is the outer bound on how long pre-cutover links
  stay live.
- `messages.fileid` exists only in MySQL. The code doesn't use it and it is NULL in every
  row. `load` adds it temporarily, loads, asserts it is still all NULL, then drops it.
- `uniqueid` and upload ids are mixed-case. MySQL compared them case-insensitively
  (latin1/utf8mb4 `_ci` collations); Postgres compares case-sensitively. Links carry the
  exact generated case, so only a hand-retyped link with the wrong case would stop
  resolving. The ids are distinct under both rules.
- Slack tokens never rotate: 0 refresh tokens and 0 expiries. Tokens stay valid in
  MySQL after cutover, so a rollback doesn't break Slack.
- `information_schema.tables.AUTO_INCREMENT` is cached for a day
  (`information_schema_stats_expiry` = 86400). `sequences.py` reads it with the cache
  turned off.
- Sequences continue from `AUTO_INCREMENT`, not `MAX(id)`: messages → 2647,
  reminders → 40, Slack → 12. Ids of deleted rows are never issued again.

## Rehearsal (2026-11-16)

Ran against live prod MySQL into throwaway `passwordexchange_rehearsal` and
`slack_oauth_rehearsal` databases. The full sequence ran top to bottom from empty
databases, followed by a `reset` → `load` → `verify` retry.

| stage | time |
|---|---|
| preflight (includes the 30 s checksum wait) | 33 s |
| prepare (migrations + Slack tables) | 6 s |
| load (3 pgloader passes under amd64 emulation, sequences, schema diff) | 55 s |
| verify (row diff + Slack lookups) | 6 s |

Results:
- All 5 tables matched row for row, with matching digests and every sequence past the source max.
- Both Slack teams resolved through the slackbot store.
- Negative control: changing one cell in the target made `verify` report `MISMATCH … [view_count]` and exit 1.

Things the rehearsal caught, now fixed in the tooling:
- pgloader maps the MySQL database name to a Postgres schema, so it needs `ALTER SCHEMA … RENAME TO 'public'`.
- In data-only mode pgloader **drops target foreign keys and never recreates them**, so the load files use `no foreign keys`. The schema-diff gate exists to catch anything similar.
- `SET MySQL PARAMETERS time_zone = '+00:00'` fails through pgloader's MySQL driver. The server is UTC (gated) instead.
- The MySQL users can see other apps' threads on the shared instance, so the freeze check filters by user.

The data volume is tiny. Expect the write freeze to last **≈15 min**, mostly Kubernetes
scale-down and scale-up, not copying.

## Setup (any time before the window)

```bash
# Operator laptop. ADC must be valid: gcloud auth application-default login
W=$(mktemp -d) && chmod 700 "$W"   # holds creds, load files, dumps, logs
podman pod create --name pgmig -p 127.0.0.1:13306:3306 -p 127.0.0.1:15432:5432
ADC=$HOME/.config/gcloud/application_default_credentials.json
podman run -d --pod pgmig --name mysqlproxy -v "$ADC:/adc.json:ro" -e GOOGLE_APPLICATION_CREDENTIALS=/adc.json \
  gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.22.0 --address 0.0.0.0 --port 3306 gen-lang-client-0479899208:us-central1:prod
podman run -d --pod pgmig --name pgproxy -v "$ADC:/adc.json:ro" -e GOOGLE_APPLICATION_CREDENTIALS=/adc.json \
  gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.22.0 --address 0.0.0.0 --port 5432 gen-lang-client-0479899208:us-central1:sre-bible

python3 -m venv "$W/venv" && "$W/venv/bin/pip" install -r deploy/migration/requirements.txt
(cd app && go build -o "$W/pxapp" .)       # from the tag being deployed
```

Creds file `$W/creds.env` (mode 0600). Fill it without echoing values:
- `MY_APP_USER`, `MY_APP_PASS`, `MY_APP_DB`, `MY_OAUTH_USER`, `MY_OAUTH_PASS`, `MY_OAUTH_DB`: the current MySQL values from the live `password-exchange-db-prod` Secret.
- `PG_APP_PASS`, `PG_SLACK_PASS`: from `terraform output -json db_passwords` in `deploy/terraform` (keys `passwordexchange` and `slackbot`).

```bash
source "$W/creds.env"
export APP_BIN="$W/pxapp" PY="$W/venv/bin/python" WORK="$W"
export SLACK_CLIENT_ID=<client id of the installations>   # not secret, but keep it out of public issues
```

pgloader image: `ghcr.io/dimitri/pgloader@sha256:bd81512667c987c516c87c397d1039ce8e1a15f128e51b4cfe6177137489f4f5`
(amd64 only, so it runs under emulation on Apple silicon). It handles MySQL
`mysql_native_password` and Postgres SCRAM through the proxies.

## T-1 day

1. Check that no migration has merged since the rehearsal. If one has, `prepare` will
   fail on the version gate: rehearse again first.
2. Rehearse once more on the same laptop, into the `_rehearsal` databases (create them as
   the app users, drop them afterwards):
   `for s in preflight prepare load verify; do deploy/migration/copy.sh $s _rehearsal || break; done`
3. Prepare both Secret variants as files. Never print them.
   ```bash
   umask 077; CTX=gke_gen-lang-client-0479899208_us-central1-f_prod
   kubectl --context $CTX -n default get secret password-exchange-db-prod -o json \
     | jq '{apiVersion,kind,type,metadata:{name:.metadata.name,namespace:.metadata.namespace},data}' > "$W/secret-mysql.json"
   PW=$(cd deploy/terraform && terraform output -json db_passwords)
   jq --argjson pw "$PW" '.data += ({
       PASSWORDEXCHANGE_DBDRIVER:"postgres", PASSWORDEXCHANGE_DBHOST:"127.0.0.1", PASSWORDEXCHANGE_DBPORT:"5432",
       PASSWORDEXCHANGE_DBNAME:"passwordexchange", PASSWORDEXCHANGE_DBUSER:"passwordexchange",
       PASSWORDEXCHANGE_DBPASS:$pw.passwordexchange,
       OAUTHDB_USER:"slackbot", OAUTHDB_NAME:"slack_oauth", OAUTHDB_PASSWORD:$pw.slackbot
     } | map_values(@base64))' "$W/secret-mysql.json" > "$W/secret-postgres.json"
   # review the non-secret keys only
   jq -r '.data|to_entries[]|select(.key|test("PASS")|not)|"\(.key)=\(.value|@base64d)"' "$W/secret-postgres.json"
   ```
   `PASSWORDEXCHANGE_AUTOMIGRATE` stays `false`, `DBMAXOPENCONNS` stays 2 and the
   OAuth pool stays 1/0. Every pod already has the `cloud-sql-proxy` sidecar on
   127.0.0.1:5432 (#568).
4. Check that `aws login` for the `sopsuser` profile works (`sops -d` on
   `kubernetes/password-exchange-db-prod.encrypted.yaml`). It's needed after cutover.
5. Tell the maintainer the window. No merges to master from the start of the freeze
   until the end of step 9.

## Cutover

`CTX` as above. Note the start time; every step has a go/no-go.

1. **Stop dev auto-sync.** `passwordexchange-devv` owns `slackbot-deployment` and would scale it back up on any sync.
   `kubectl --context $CTX -n argocd patch application passwordexchange-devv --type merge -p '{"spec":{"syncPolicy":{"automated":null}}}'`
2. **Freeze writers:**
   ```bash
   kubectl --context $CTX -n default patch cronjob password-exchange-reminder-prod -p '{"spec":{"suspend":true}}'
   kubectl --context $CTX -n default scale deploy/database-prod-deployment deploy/slackbot-deployment --replicas=0
   kubectl --context $CTX -n default wait --for=delete pod -l app=database,phase=prod --timeout=120s
   kubectl --context $CTX -n default wait --for=delete pod -l app=slackbot --timeout=120s
   ```
   `phase=prod` matters: the dev database pods also carry `app=database`. "no matching resources"
   means the pods are already gone. Step 4's session gate is the real proof that no writer is left.
   The web front end stays up and returns errors on submit/view until step 7. Any
   reminder Job already running must finish first: wait until
   `kubectl --context $CTX -n default get pods -l component=reminder,phase=prod` shows none running.
3. **Rollback artifact.** Take a consistent dump of the frozen source, kept 0600 in `$W` only:
   ```bash
   MYSQL_PWD="$MY_APP_PASS"   mysqldump --skip-ssl -h127.0.0.1 -P13306 -u"$MY_APP_USER"   --single-transaction --no-tablespaces "$MY_APP_DB"   > "$W/rollback-app.sql"
   MYSQL_PWD="$MY_OAUTH_PASS" mysqldump --skip-ssl -h127.0.0.1 -P13306 -u"$MY_OAUTH_USER" --single-transaction --no-tablespaces "$MY_OAUTH_DB" > "$W/rollback-oauth.sql"
   ```
4. **`copy.sh preflight`**. Go when the time zone is UTC, both users show **0** other
   sessions and the checksums are stable 30 s apart. If sessions are non-zero, find who
   holds them before going on: something outside password-exchange may be using these
   users.
5. **`copy.sh prepare`**. Go when both targets were empty, the migration versions match
   and the Slack tables were created.
6. **`copy.sh load`, then `copy.sh verify`**. Go on `RESULT: PASS` and `SLACK LOOKUPS: PASS`.
   If `load` is interrupted (laptop, proxy, Spot preemption), run `copy.sh reset`, then
   `load` again.
7. **Switch:**
   ```bash
   kubectl --context $CTX apply -f "$W/secret-postgres.json"
   kubectl --context $CTX -n default scale deploy/database-prod-deployment --replicas=2
   kubectl --context $CTX -n default scale deploy/slackbot-deployment --replicas=2
   kubectl --context $CTX -n default rollout status deploy/database-prod-deployment deploy/slackbot-deployment
   kubectl --context $CTX -n default patch cronjob password-exchange-reminder-prod -p '{"spec":{"suspend":false}}'
   ```
8. **Smoke test.**
   - `https://password.exchange` health shows the expected version.
   - Create a message, open it once and check it is gone after the view.
   - The next reminder Job (≤10 min) completes.
   - The database pods' logs show no MySQL host.
   - A human runs `/password` in Slack in both workspaces.
   - Do **not** open a pre-cutover link to test it: views are counted.
9. **Re-enable dev auto-sync** (prune on, no self-heal, as before):
   `kubectl --context $CTX -n argocd patch application passwordexchange-devv --type merge -p '{"spec":{"syncPolicy":{"automated":{"prune":true}}}}'`.
   Merges may resume.
10. **Commit the Secret.** Re-encrypt `kubernetes/password-exchange-db-prod.encrypted.yaml`
    from `secret-postgres.json` (SOPS, profile `sopsuser`) and open a PR. Keep
    `secret-mysql.json` for the rollback window.

**Spot preemption.** The node pool is all Spot. A preemption can take out kube-dns, the
service mesh or the ArgoCD repo-server for a few minutes. If a `kubectl` or `argocd` step
times out, wait for those to recover, then repeat the step. Steps 2, 7 and 9 are
idempotent, and `copy.sh` stages can be rerun (`load` via `reset`).

## Rollback (within 24 h)

```bash
kubectl --context $CTX -n default patch cronjob password-exchange-reminder-prod -p '{"spec":{"suspend":true}}'
kubectl --context $CTX -n default scale deploy/database-prod-deployment deploy/slackbot-deployment --replicas=0
kubectl --context $CTX apply -f "$W/secret-mysql.json"
kubectl --context $CTX -n default scale deploy/database-prod-deployment deploy/slackbot-deployment --replicas=2
kubectl --context $CTX -n default patch cronjob password-exchange-reminder-prod -p '{"spec":{"suspend":false}}'
```

MySQL is never written after cutover, so rolling back returns to the frozen state of
step 3. **What a rollback costs:**
- Messages created on Postgres are lost, and their links stop working.
- Messages viewed or deleted on Postgres come back from MySQL with their pre-cutover
  view counts. A view-once secret could be opened a second time.
- New Slack installs made after cutover are lost. Existing tokens keep working because
  they never rotate.

So, for a password-sharing service, roll back only for a fault that blocks the service.
For anything less, fix forward. (`rollback-*.sql` is a last-resort copy of the frozen
source if MySQL itself is damaged.)

## After the 24 h window

1. Show that nothing else uses the MySQL users. The per-user session counts from
   `copy.sh preflight` must stay 0 for a day. Also check the configuration of every
   workload that reads the shared `test-secret`, since several apps outside
   password-exchange load all of its keys.
2. Revoke writes: `REVOKE INSERT, UPDATE, DELETE ON passwordExchange.* FROM 'passwordexchange'@'%'`, and the same for `oauth.*` and `oauthUser`. Run it as an admin user.
3. Delete `$W`: creds, Secret files, rollback dumps, load logs and schema dumps. Then
   `podman pod rm -f pgmig`. Record that it was done on #569.
4. Close #569. Dropping the MySQL databases and users is #570.

## Rehearsal cleanup log

- 2026-11-16: dropped `passwordexchange_rehearsal` and `slack_oauth_rehearsal`. They held
  real Slack tokens and upload keys. Deleted the local load files, logs and creds; stopped
  the `pgmig` pod.
