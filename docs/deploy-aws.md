# Deploying to AWS

This runs the web app and API as separate containers on ECS Fargate behind an Application Load Balancer, with the database and Redis outside Docker. It hasn't been run against a real AWS account yet, so expect to adjust details on the first deploy.

```text
              ┌─ app.example.com ─▶ web (ECS, port 3000) ─┐
Browser ─▶ ALB                                            ├─ internal ─▶ api
              └─ api.example.com ─▶ api (ECS, port 5000) ◀┘
                                         │
                       Neon or RDS ◀─────┴─────▶ ElastiCache or Upstash
```

## What you need

- Two ECR repositories, such as `myapp-web` and `myapp-api`
- A VPC with an ALB, an ACM certificate for your domain, and an ECS cluster
- Postgres on Neon or RDS
- Redis on ElastiCache or Upstash, which is required for rate limiting

The web app and API must share a parent domain, because auth cookies use `SameSite=Lax`. For example `app.example.com` and `api.example.com`, with `COOKIE_DOMAIN=.example.com`.

## 1. Build and push the images

```sh
aws ecr get-login-password | docker login --username AWS --password-stdin "$REGISTRY"

docker build -t "$REGISTRY/myapp-api:$TAG" apps/api
docker push "$REGISTRY/myapp-api:$TAG"

docker build -f apps/web/Dockerfile \
  --build-arg NEXT_PUBLIC_API_URL=https://api.example.com \
  -t "$REGISTRY/myapp-web:$TAG" .
docker push "$REGISTRY/myapp-web:$TAG"
```

Next builds `NEXT_PUBLIC_API_URL` into the web image, so each environment with a different API URL needs its own build. Build for the CPU architecture of your tasks, using `--platform linux/amd64` or `linux/arm64`.

## 2. Create the task definitions

**API:** container port 5000.

| Variable | Value | Store in |
| --- | --- | --- |
| `DATABASE_URL` | See [Database](#database) | Secrets Manager |
| `MIGRATE_DATABASE_URL` | Direct connection. See [Database](#database) | Secrets Manager |
| `REDIS_URL` | `rediss://...` | Secrets Manager |
| `JWT_ACCESS_SECRET` | `openssl rand -base64 48` | Secrets Manager |
| `JWT_REFRESH_SECRET` | `openssl rand -base64 48`, a different value | Secrets Manager |
| `SMTP_USERNAME` | From your email provider. See [Email](#email) | Secrets Manager |
| `SMTP_PASSWORD` | From your email provider | Secrets Manager |
| `SMTP_HOST` | Such as `email-smtp.us-east-1.amazonaws.com` | Task definition |
| `SMTP_PORT` | `587` | Task definition |
| `EMAIL_FROM` | Such as `no-reply@example.com` | Task definition |
| `FRONTEND_URL` | `https://app.example.com` | Task definition |
| `COOKIE_DOMAIN` | `.example.com` | Task definition |
| `TRUSTED_PROXIES` | Your VPC CIDR, such as `10.0.0.0/16` | Task definition |

The API refuses to start if the JWT secrets are weak, or `FRONTEND_URL`, `COOKIE_DOMAIN`, `SMTP_HOST` or `EMAIL_FROM` are missing.

`TRUSTED_PROXIES` matters behind an ALB. If it's unset, every request appears to come from the load balancer, so all users share one rate limit. Don't set it wider than your VPC, or clients can fake their IP.

**Web:** container port 3000.

| Variable | Value |
| --- | --- |
| `INTERNAL_API_URL` | How the web container reaches the API, such as `http://api:5000` with Service Connect |
| `COOKIE_DOMAIN` | `.example.com` |

Use ECS Service Connect for `INTERNAL_API_URL`, so server-side requests stay inside the VPC. `https://api.example.com` also works, but goes through the load balancer.

## 3. Create the load balancer rules

| Host | Target | Health check path |
| --- | --- | --- |
| `app.example.com` | Web service, port 3000 | `/api/health` |
| `api.example.com` | API service, port 5000 | `/api/v1/health` |

Terminate HTTPS at the ALB and redirect HTTP to HTTPS. Auth cookies are marked `Secure`, so login doesn't work over plain HTTP.

Point the ALB at `/api/v1/health`, not `/api/v1/ready`. `/ready` fails when the database is unreachable, and the ALB would then replace every healthy container during a brief database problem. Use `/ready` for monitoring and alerts.

## 4. Run migrations

Run these as a one-off task before starting new API tasks. It uses the API image with a different command:

```sh
aws ecs run-task --cluster "$CLUSTER" --launch-type FARGATE \
  --task-definition myapp-api \
  --network-configuration "awsvpcConfiguration={subnets=[$SUBNET],securityGroups=[$SECURITY_GROUP]}" \
  --overrides '{"containerOverrides":[{"name":"api","command":["/app/migrate","up"]}]}'
```

`/app/migrate down N` rolls back the last N migrations, and `/app/migrate version` prints the current version.

## 5. Start the services

Create the ECS services and attach them to the target groups. On later deploys, run migrations first, then update the services to the new image tag. Keep migrations backward compatible, since old and new tasks run side by side during a rolling deploy.

To create a first admin, run `/app/seed` the same way, with `SEED_ADMIN_PASSWORD` and `SEED_MEMBER_PASSWORD` set as real values. It refuses `change-me` passwords.

## Database

The app uses pgx, so settings go in the connection string.

**Neon**
- `DATABASE_URL`: the pooled connection string (host contains `-pooler`), with `sslmode=require`.
- `MIGRATE_DATABASE_URL`: the direct connection string. Migrations hold locks, which don't work through the pooler.
- If you see prepared statement errors through the pooler, add `default_query_exec_mode=exec` to `DATABASE_URL`. This hasn't been confirmed against a real Neon database yet.

**RDS**
- Use the instance endpoint for both, with `sslmode=require`. `MIGRATE_DATABASE_URL` can stay unset.
- `sslmode=require` encrypts the connection but doesn't verify the server. For `verify-full`, add the [RDS CA bundle](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.SSL.html) to the image and set `sslrootcert`.
- Put RDS in the VPC, and allow port 5432 only from the API's security group. Also allow it from the migration task if that uses a different group.

The default pool is the larger of 4 connections or the number of CPUs, per container. To change it, add `pool_max_conns=10` to the connection string, and keep the total across all tasks below your database's limit.

## Email

Verification and password reset emails go over SMTP, so any provider works. New users can't log in until they verify their email, so working email is required before anyone can sign up. TLS is required in production. These steps are for Amazon SES:

1. Verify your domain in the SES console, and add the DNS records it gives you. This also sets up DKIM.
2. Request production access. New SES accounts are in a sandbox that can only send to verified addresses.
3. Create SMTP credentials in the SES console, under **SMTP settings**. They aren't the same as your AWS access keys.
4. Set `SMTP_HOST` to the endpoint for your region, `SMTP_PORT=587`, and `EMAIL_FROM` to an address on the verified domain.

API tasks need outbound access to the SMTP endpoint on port 587, through a NAT gateway or VPC endpoint if they run in private subnets. If email can't be sent, the API logs the error and still returns its normal response, so set an alert on the failures, described below.

This setup hasn't been tested against real SES yet.

### Alert when email fails

Users see the same "check your email" message even when sending fails, so nothing tells you it's broken unless you watch the logs. Production logs are JSON, which CloudWatch can filter on. Create a metric from the failures, and an alarm on the metric:

```sh
aws logs put-metric-filter \
  --log-group-name "$API_LOG_GROUP" \
  --filter-name account-email-failures \
  --filter-pattern '{ ($.msg = "failed to send password*") || ($.msg = "failed to send verification*") }' \
  --metric-transformations metricName=AccountEmailFailures,metricNamespace=MyApp,metricValue=1,defaultValue=0

aws cloudwatch put-metric-alarm \
  --alarm-name account-email-failures \
  --namespace MyApp --metric-name AccountEmailFailures \
  --statistic Sum --period 300 --evaluation-periods 1 --threshold 1 \
  --comparison-operator GreaterThanOrEqualToThreshold \
  --treat-missing-data notBreaching \
  --alarm-actions "$SNS_TOPIC_ARN"
```

These commands haven't been run against a real AWS account yet.

## Security groups

| Resource | Allow inbound from |
| --- | --- |
| ALB | Internet, ports 80 and 443 |
| Web tasks | ALB, port 3000 |
| API tasks | ALB and web tasks, port 5000 |
| RDS | API tasks, port 5432 |
| ElastiCache | API tasks, port 6379 |

## Test the images before deploying

`docker-compose.prod.yml` runs the production images on your machine:

```sh
cp .env.production.example .env.production   # then fill in the JWT secrets
docker compose --env-file .env.production -f docker-compose.prod.yml --profile local-deps up --build
docker compose --env-file .env.production -f docker-compose.prod.yml run --rm api /app/migrate up
```

For local testing, set `FRONTEND_URL` and `NEXT_PUBLIC_API_URL` to your `localhost` URLs, and `COOKIE_DOMAIN=localhost`. Emails go to Mailpit at <http://localhost:8025>.

## Running without Docker

Each app also builds and runs on its own:

```sh
# API: a single binary. Set the variables above in the environment.
cd apps/api && make build && GIN_MODE=release ./server

# Web
bun run build --filter=web
bun run start --filter=web
```
