create table public.device_accounts (
  id uuid primary key references auth.users(id) on delete cascade,
  device_identifier_hash text not null unique
    check (device_identifier_hash ~ '^[0-9a-f]{64}$'),
  access_code_hash text not null unique
    check (access_code_hash ~ '^[0-9a-f]{64}$'),
  access_code_counter integer not null
    check (access_code_counter >= 0),
  login_email text not null unique,
  created_at timestamptz not null default now(),
  last_login_at timestamptz
);

comment on table public.device_accounts is
  'Private mapping between device identifiers, access codes, and Supabase Auth users.';
comment on column public.device_accounts.device_identifier_hash is
  'HMAC-SHA-256 digest; the raw device identifier is never stored.';
comment on column public.device_accounts.access_code_hash is
  'HMAC-SHA-256 digest used to look up an access code during login.';
comment on column public.device_accounts.access_code_counter is
  'Deterministic derivation counter used when MOD 11-2 produces X or a collision.';

alter table public.device_accounts enable row level security;
revoke all on table public.device_accounts from anon, authenticated;
grant select, insert, update, delete
  on table public.device_accounts
  to service_role;

create table public.scores (
  user_id uuid primary key references public.device_accounts(id) on delete cascade,
  score bigint not null
    check (score between 0 and 9007199254740991),
  max_combo bigint not null default 0
    check (max_combo between 0 and 9007199254740991),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index scores_ranking_idx
  on public.scores (score desc, updated_at asc);

alter table public.scores enable row level security;

create policy "Scores are publicly readable"
  on public.scores
  for select
  to anon, authenticated
  using (true);

revoke all on table public.scores from anon, authenticated;
grant select, insert, update, delete
  on table public.scores
  to service_role;
grant select on table public.scores to anon, authenticated;

create or replace function public.set_score_updated_at()
returns trigger
language plpgsql
set search_path = ''
as $$
begin
  new.updated_at = now();
  return new;
end;
$$;

revoke all on function public.set_score_updated_at() from public;
grant execute on function public.set_score_updated_at() to service_role;

create trigger set_score_updated_at
before update on public.scores
for each row
execute function public.set_score_updated_at();

create view public.leaderboard
with (security_invoker = true)
as
select
  dense_rank() over (order by score desc) as rank,
  user_id,
  score,
  max_combo,
  updated_at
from public.scores;

revoke all on table public.leaderboard from anon, authenticated;
grant select on table public.leaderboard to anon, authenticated;

comment on view public.leaderboard is
  'Public score ranking without exposing device or access-code metadata.';
