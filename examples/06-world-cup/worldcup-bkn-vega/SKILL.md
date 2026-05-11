# 世界杯 BKN（Fjelstul · Vega `resource` 占位符版）

本目录为本示例 **离线 BKN 模版**：对象类数据源为 **`resource | {{*_RES_ID}}`**（Vega table Resource UUID）。流程见 **`README.zh.md`**、**`../WORKFLOW-BRANCH-VEGA.zh.md`**。

**关系 ID**：`rel_<表名>_<外键列去掉_id>`（约 **29** 条「minimal」骨架边；历史上由生成脚本按 CSV FK 推导，本树为检入快照）。

## 校验

`mkdir -p .tmp && TMPDIR="$(pwd)/.tmp" kweaver bkn validate ./worldcup-bkn-vega`

## 对象类型

| ID | Name | Path |
|----|------|------|
| `award_winners` | 奖项得主 | `object_types/award_winners.bkn` |
| `awards` | 奖项定义 | `object_types/awards.bkn` |
| `bookings` | 黄牌红牌记录 | `object_types/bookings.bkn` |
| `confederations` | 足球联合会 | `object_types/confederations.bkn` |
| `goals` | 进球事件 | `object_types/goals.bkn` |
| `group_standings` | 小组积分榜 | `object_types/group_standings.bkn` |
| `groups` | 小组 | `object_types/groups.bkn` |
| `host_countries` | 东道主国家 | `object_types/host_countries.bkn` |
| `manager_appearances` | 教练临场指挥记录 | `object_types/manager_appearances.bkn` |
| `manager_appointments` | 主帅执教任命 | `object_types/manager_appointments.bkn` |
| `managers` | 主教练 | `object_types/managers.bkn` |
| `matches` | 比赛场次 | `object_types/matches.bkn` |
| `penalty_kicks` | 点球事件 | `object_types/penalty_kicks.bkn` |
| `player_appearances` | 球员每场比赛出场 | `object_types/player_appearances.bkn` |
| `players` | 球员 | `object_types/players.bkn` |
| `qualified_teams` | 出线球队资格 | `object_types/qualified_teams.bkn` |
| `referee_appearances` | 裁判执法场次 | `object_types/referee_appearances.bkn` |
| `referee_appointments` | 裁判指派 | `object_types/referee_appointments.bkn` |
| `referees` | 裁判员 | `object_types/referees.bkn` |
| `squads` | 阵容名单 | `object_types/squads.bkn` |
| `stadiums` | 比赛球场 | `object_types/stadiums.bkn` |
| `substitutions` | 换人事件 | `object_types/substitutions.bkn` |
| `team_appearances` | 球队每届参赛记录 | `object_types/team_appearances.bkn` |
| `teams` | 国家队 | `object_types/teams.bkn` |
| `tournament_stages` | 赛事阶段 | `object_types/tournament_stages.bkn` |
| `tournament_standings` | 赛事总积分榜 | `object_types/tournament_standings.bkn` |
| `tournaments` | 世界杯赛事 | `object_types/tournaments.bkn` |

## 关系类型（语义化 ID）

| Relation ID | 名称（节选） | Path |
|-------------|-------------|------|
| `rel_award_winners_award` | 奖项得主·奖项→奖项定义 | `relation_types/rel_award_winners_award.bkn` |
| `rel_award_winners_player` | 奖项得主·球员→球员 | `relation_types/rel_award_winners_player.bkn` |
| `rel_award_winners_team` | 奖项得主·球队→国家队 | `relation_types/rel_award_winners_team.bkn` |
| `rel_award_winners_tournament` | 奖项得主·届次赛事→世界杯赛事 | `relation_types/rel_award_winners_tournament.bkn` |
| `rel_bookings_match` | 黄牌红牌记录·场次→比赛场次 | `relation_types/rel_bookings_match.bkn` |
| `rel_bookings_player` | 黄牌红牌记录·球员→球员 | `relation_types/rel_bookings_player.bkn` |
| `rel_bookings_team` | 黄牌红牌记录·球队→国家队 | `relation_types/rel_bookings_team.bkn` |
| `rel_goals_match` | 进球事件·场次→比赛场次 | `relation_types/rel_goals_match.bkn` |
| `rel_goals_player` | 进球事件·球员→球员 | `relation_types/rel_goals_player.bkn` |
| `rel_goals_team` | 进球事件·球队→国家队 | `relation_types/rel_goals_team.bkn` |
| `rel_group_standings_team` | 小组积分榜·球队→国家队 | `relation_types/rel_group_standings_team.bkn` |
| `rel_group_standings_tournament` | 小组积分榜·届次赛事→世界杯赛事 | `relation_types/rel_group_standings_tournament.bkn` |
| `rel_groups_tournament` | 小组·届次赛事→世界杯赛事 | `relation_types/rel_groups_tournament.bkn` |
| `rel_matches_away_team` | 比赛场次·客队→国家队 | `relation_types/rel_matches_away_team.bkn` |
| `rel_matches_home_team` | 比赛场次·主队→国家队 | `relation_types/rel_matches_home_team.bkn` |
| `rel_matches_stadium` | 比赛场次·球场→比赛球场 | `relation_types/rel_matches_stadium.bkn` |
| `rel_matches_tournament` | 比赛场次·届次赛事→世界杯赛事 | `relation_types/rel_matches_tournament.bkn` |
| `rel_player_appearances_match` | 球员每场比赛出场·场次→比赛场次 | `relation_types/rel_player_appearances_match.bkn` |
| `rel_player_appearances_player` | 球员每场比赛出场·球员→球员 | `relation_types/rel_player_appearances_player.bkn` |
| `rel_player_appearances_team` | 球员每场比赛出场·球队→国家队 | `relation_types/rel_player_appearances_team.bkn` |
| `rel_qualified_teams_team` | 出线球队资格·球队→国家队 | `relation_types/rel_qualified_teams_team.bkn` |
| `rel_qualified_teams_tournament` | 出线球队资格·届次赛事→世界杯赛事 | `relation_types/rel_qualified_teams_tournament.bkn` |
| `rel_squads_player` | 阵容名单·球员→球员 | `relation_types/rel_squads_player.bkn` |
| `rel_squads_team` | 阵容名单·球队→国家队 | `relation_types/rel_squads_team.bkn` |
| `rel_squads_tournament` | 阵容名单·届次赛事→世界杯赛事 | `relation_types/rel_squads_tournament.bkn` |
| `rel_team_appearances_match` | 球队每届参赛记录·场次→比赛场次 | `relation_types/rel_team_appearances_match.bkn` |
| `rel_team_appearances_opponent` | 球队每届参赛记录·对手球队→国家队 | `relation_types/rel_team_appearances_opponent.bkn` |
| `rel_team_appearances_team` | 球队每届参赛记录·球队→国家队 | `relation_types/rel_team_appearances_team.bkn` |
| `rel_teams_confederation` | 国家队·联合会→足球联合会 | `relation_types/rel_teams_confederation.bkn` |

共 **29** 条关系。

## 概念分组

| ID | Path |
|----|------|
| worldcup_competition | `concept_groups/worldcup_competition.bkn` |
| worldcup_participants | `concept_groups/worldcup_participants.bkn` |
| worldcup_events | `concept_groups/worldcup_events.bkn` |
| worldcup_honors | `concept_groups/worldcup_honors.bkn` |
