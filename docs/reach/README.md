# Daily reach records

The owner's standing instruction (2026-11-18): **check what we published every
day**. This directory is where that check lands.

- `latest.md` — the most recent run, overwritten each time.
- `YYYY-MM-DD.md` — that day's runs, appended, so a morning and an evening check
  both survive.

## Automatic half

```sh
node tools/daily-reach.mjs            # write latest.md + today's file
node tools/daily-reach.mjs --stdout   # print, write nothing
```

It collects, and labels the window for each:

- GitHub views / clones / referrers / stars, from `tools/traffic-report.mjs`
  (same numbers, same caveats as [docs/METRICS.md](../METRICS.md)).
- Material downloads per release, all time (subtract the ~1–2/project per release
  that is our own verification before calling any of it a user).
- npm weekly totals and the **latest share** of the newest version present in the
  window. Never quote the totals as users: the crawler fleet inflates them
  (a fresh release takes a day or two to appear per version).
- Reddit post JSON (score, comments, `removed_by_category`) when this machine can
  reach reddit.com. On campus Wi-Fi it cannot; the report says `unreachable` and
  the browser check below covers it.
- dev.to articles by the account (`https://dev.to/api/articles?username=…`):
  reactions and comment counts, public and unauthenticated.
- Upstream PR state: `awesome-dsh-plugin` tarball bump and the winget package.

A Windows Scheduled Task, `AITerminalPet-DailyReach`, runs it daily at 09:00 and
at the next logon if the machine was off. Check it with:

```powershell
Get-ScheduledTask -TaskName AITerminalPet-DailyReach
```

## Manual half (needs the signed-in browser)

The script cannot see these, and pretending otherwise would be worse than a
checklist:

1. **X** — open the thread and `/notifications`; record likes, follows and
   replies. The thread view's "1 reply" is our own second post.
2. **Reddit** — read the comments as the author; a removed post can still show
   on the profile. `r/SideProject` stays filtered until a moderator replies.
3. **dev.to** — open the articles and read the comments; the API gives counts,
   not the text. The launch retrospective is at
   <https://dev.to/jjjjjjjjnnjnn/19-visitors-1-star-1416-clones-what-a-launch-actually-looked-like-43pi>.
4. **HN / V2EX** — not posted; the body must be written and submitted by the
   owner, not generated.
5. **Replies** — write them into that day's file under the relevant post, with
   the author and the substance. A reply that is not recorded will be forgotten
   by the next session.

## Rules that did not change

- Never call npm totals or release downloads "users".
- Never quote a number without its window.
- Do not repost a filtered Reddit post, and do not ask for votes.
- A quiet day is a valid result. Write "no change" instead of inventing activity.

## Committing

The scheduled job only writes files. Commit them with the next session's work,
or weekly if screenshots/notes were added by hand; do not commit a file per run.
