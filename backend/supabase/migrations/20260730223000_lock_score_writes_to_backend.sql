drop policy if exists "Users can insert their own score"
  on public.scores;
drop policy if exists "Users can update their own score"
  on public.scores;

revoke insert (user_id, score, max_combo)
  on table public.scores
  from authenticated;
revoke update (score, max_combo)
  on table public.scores
  from authenticated;

grant execute on function public.set_score_updated_at()
  to service_role;
