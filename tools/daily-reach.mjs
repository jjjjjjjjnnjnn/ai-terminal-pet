#!/usr/bin/env node
/**
 * Daily reach check: what happened to the things we published, in one file.
 *
 *   node tools/daily-reach.mjs            # write docs/reach/<today>.md + latest.md
 *   node tools/daily-reach.mjs --stdout   # print only, write nothing
 *
 * Why this exists
 * ---------------
 * Promotion is a loop, not a launch: a post that got three replies and a post
 * that got none look identical the next morning unless somebody measures. The
 * machine-checkable half is here (GitHub traffic, release downloads, npm
 * versions, Reddit post JSON, upstream PR state); the half that needs a signed-in
 * browser (X replies, Reddit comments as the author, HN/V2EX posting) is printed
 * as a checklist, because inventing an API for those would be a worse lie than
 * saying "check it by hand".
 *
 * The numbers are not an audience: see docs/METRICS.md for why npm totals and
 * release downloads have baselines. This file records them with their window and
 * never calls them users.
 *
 * Exit code is 0 even when a source is unreachable. A daily job that fails
 * because Reddit is blocked on campus Wi-Fi trains people to ignore it; the
 * report says "unreachable" and the human check covers it.
 */

import { execFileSync } from "node:child_process";
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repoDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const stdoutOnly = process.argv.includes("--stdout");

const REPO = "jjjjjjjjnnjnn/ai-terminal-pet";
const DEVTO_USER = "jjjjjjjjnnjnn";
const X_THREAD = "https://x.com/JJjj1955714/status/2122943194904322144";
const REDDIT_POSTS = [
  {
    label: "r/coolgithubprojects",
    url: "https://www.reddit.com/r/coolgithubprojects/comments/1xzj9zc/ai_terminal_pet_a_56px_alwaysontop_strip_that/"
  },
  {
    label: "r/SideProject",
    url: "https://www.reddit.com/r/SideProject/comments/1xzj589/ai_terminal_pet_a_56px_alwaysontop_strip_that/"
  }
];
const UPSTREAM = [
  { label: "awesome-dsh-plugin #16703 (tarball 1.2.35)", repo: "awesome-dsh-plugin/awesome-dsh-plugin", number: 16703 },
  { label: "winget-pkgs #481818 (1.0.15)", repo: "microsoft/winget-pkgs", number: 481818 }
];

function run(command, args, options = {}) {
  try {
    return {
      ok: true,
      out: execFileSync(command, args, {
        cwd: repoDir,
        encoding: "utf8",
        timeout: options.timeout || 180000,
        stdio: ["ignore", "pipe", "pipe"]
      }).trim()
    };
  } catch (error) {
    return { ok: false, out: String((error && error.stderr) || (error && error.message) || error).trim() };
  }
}

function traffic() {
  const result = run(process.execPath, ["tools/traffic-report.mjs", "--json"]);
  if (!result.ok) return { error: result.out };
  try {
    return JSON.parse(result.out);
  } catch (error) {
    return { error: `traffic-report JSON did not parse: ${error.message}` };
  }
}

function releases() {
  const result = run("gh", ["api", `repos/${REPO}/releases?per_page=5`]);
  if (!result.ok) return { error: result.out.split("\n")[0] };
  try {
    return JSON.parse(result.out).map((release) => ({
      tag: release.tag_name,
      published: release.published_at,
      assets: release.assets.map((asset) => ({ name: asset.name, downloads: asset.download_count }))
    }));
  } catch (error) {
    return { error: `releases JSON did not parse: ${error.message}` };
  }
}

function upstream() {
  return UPSTREAM.map((item) => {
    const result = run("gh", [
      "pr", "view", String(item.number), "-R", item.repo,
      "--json", "state,mergedAt,updatedAt,title"
    ]);
    if (!result.ok) return { ...item, state: "unreachable" };
    try {
      const data = JSON.parse(result.out);
      return { ...item, state: data.state, mergedAt: data.mergedAt, updatedAt: data.updatedAt };
    } catch {
      return { ...item, state: "unreadable" };
    }
  });
}

async function devto() {
  try {
    const response = await fetch(
      `https://dev.to/api/articles?username=${DEVTO_USER}&per_page=10`,
      { headers: { accept: "application/json" }, signal: AbortSignal.timeout(10000) }
    );
    if (!response.ok) return { error: `HTTP ${response.status}` };
    const items = await response.json();
    const articles = items.map((article) => ({
      id: article.id,
      title: article.title,
      url: article.url,
      published: article.published_at,
      reactions: article.public_reactions_count,
      comments: article.comments_count,
      commentsList: []
    }));
    for (const article of articles.slice(0, 3)) {
      if (!article.comments) continue;
      try {
        const listResponse = await fetch(`https://dev.to/api/comments?a_id=${article.id}`, {
          headers: { accept: "application/json" },
          signal: AbortSignal.timeout(10000)
        });
        if (!listResponse.ok) continue;
        const list = await listResponse.json();
        article.commentsList = list.slice(0, 3).map((comment) => ({
          author: (comment.user && comment.user.name) || "unknown",
          body: String(comment.body_html || "")
            .replace(/<[^>]+>/g, " ")
            .replace(/\s+/g, " ")
            .trim()
            .slice(0, 240),
          created: comment.created_at
        }));
      } catch {
        // A comment fetch is best effort; the count still stands.
      }
    }
    return articles;
  } catch (error) {
    return { error: error.name === "TimeoutError" ? "timeout" : String(error.message || error) };
  }
}

async function redditPost(post) {  const url = post.url.replace(/\/$/, "") + "/.json?limit=50";
  try {
    const response = await fetch(url, {
      headers: {
        accept: "application/json",
        "user-agent": "ai-terminal-pet-reach/1.0 (daily check; by /u/Visible_Swim471)"
      },
      signal: AbortSignal.timeout(10000)
    });
    if (!response.ok) return { ...post, status: `HTTP ${response.status}` };
    const data = await response.json();
    const body = data[0]?.data?.children?.[0]?.data || {};
    const comments = (data[1]?.data?.children || []).filter((child) => child.kind === "t1");
    return {
      ...post,
      status: body.removed_by_category ? `removed (${body.removed_by_category})` : "visible",
      score: body.score,
      comments: comments.length,
      latest: comments.slice(0, 3).map((child) => ({
        author: child.data.author,
        body: String(child.data.body || "").replace(/\s+/g, " ").slice(0, 200),
        created: child.data.created_utc
      }))
    };
  } catch (error) {
    return { ...post, status: `unreachable (${error.name === "TimeoutError" ? "timeout" : error.message})` };
  }
}

function newestVersion(npm) {
  const versions = Object.keys(npm?.week?.versions || {}).filter((version) => /^\d+\.\d+\.\d+$/.test(version));
  versions.sort((a, b) => {
    const [a1, a2, a3] = a.split(".").map(Number);
    const [b1, b2, b3] = b.split(".").map(Number);
    return a1 - b1 || a2 - b2 || a3 - b3;
  });
  return versions[versions.length - 1] || null;
}

function latestShare(npm) {
  const newest = newestVersion(npm);
  if (!newest) return null;
  const downloads = npm.week.versions[newest] || 0;
  const total = npm.week.total || 0;
  return { version: newest, downloads, total, share: total ? downloads / total : 0 };
}

function localVersion() {
  try {
    const source = readFileSync(path.join(repoDir, "internal", "app", "version.go"), "utf8");
    return /^const Version = "(\d+\.\d+\.\d+)"/m.exec(source)?.[1] || "unknown";
  } catch {
    return "unknown";
  }
}

function publishedVersion(name) {
  try {
    return JSON.parse(readFileSync(path.join(repoDir, "packages", name, "package.json"), "utf8")).version;
  } catch {
    return "?";
  }
}

function markdown(report) {
  const lines = [];
  const now = new Date();
  lines.push(`# Reach — ${now.toISOString().slice(0, 10)} (run ${now.toISOString().slice(11, 16)} UTC)`, "");
  lines.push(`Overlay ${report.localVersion} · generated by \`node tools/daily-reach.mjs\`.`, "");

  const t = report.traffic;
  if (t.error) {
    lines.push("## Traffic", "", `unavailable: ${t.error}`, "");
  } else {
    lines.push("## GitHub", "");
    lines.push(`- stars **${t.repo.stars}** · forks ${t.repo.forks} · watchers ${t.repo.subscribers} · open issues/PRs ${t.repo.issues}`);
    lines.push(`- views **${t.repo.views.count}** (${t.repo.views.uniques} unique) / 14 days`);
    lines.push(`- clones ${t.repo.clones.count} (${t.repo.clones.uniques} unique); ${t.repo.clones.ciClones} are our CI (${t.repo.clones.ciRuns} runs × 4 jobs)`);
    if (t.repo.referrers?.length) {
      lines.push(`- referrers: ${t.repo.referrers.slice(0, 5).map((r) => `${r.referrer} ${r.count}`).join(" · ")}`);
    }
    lines.push("");

    lines.push("## npm", "");
    lines.push("| package | published | week | newest in window | newest share |");
    lines.push("|---|---|---|---|---|");
    for (const [name, npm] of Object.entries(t.npm || {})) {
      const share = latestShare(npm);
      lines.push(
        `| \`${name}\` | ${publishedVersion(name)} | ${npm.week?.total ?? "?"} | ${share?.version ?? "?"} | ` +
          (share ? `${share.downloads}/${share.total} = **${(share.share * 100).toFixed(1)}%**` : "?") +
          " |"
      );
    }
    lines.push("");
    lines.push("npm's per-version windows lag a fresh release by a day or two; the share is computed from the newest version **present in the window**, not from `published`.", "");
    lines.push(`Reference \`${t.reference?.name}\`: newest share **${t.reference ? (t.reference.share * 100).toFixed(1) : "?"}%** (humans concentrate; see docs/METRICS.md).`, "");
    lines.push("## Installs (all releases)", "");
    for (const group of t.assets?.grouped || []) {
      lines.push(`- ${group.name}: **${group.downloads}** over ${group.releases} release(s) = ${group.perRelease.toFixed(1)}/release`);
    }
    lines.push("");
    lines.push("## Catalogue", "");
    if (t.catalog) {
      lines.push(`- \`${t.catalog.name}\`: ${t.catalog.downloads} downloads, rank **${t.catalog.rank}/${t.catalog.entries}** (median ${t.catalog.median})`);
    }
    lines.push("");
  }

  lines.push("## Releases (last 5)", "");
  if (report.releases.error) {
    lines.push(`unavailable: ${report.releases.error}`, "");
  } else {
    lines.push("| tag | asset downloads |");
    lines.push("|---|---|");
    for (const release of report.releases) {
      lines.push(`| \`${release.tag}\` | ${release.assets.map((a) => `${a.name} ${a.downloads}`).join("<br>")} |`);
    }
    lines.push("");
  }

  lines.push("## dev.to", "");
  if (report.devto.error) {
    lines.push(`unavailable: ${report.devto.error}`, "");
  } else if (report.devto.length === 0) {
    lines.push("no articles under this username", "");
  } else {
    lines.push("| article | published | reactions | comments |");
    lines.push("|---|---|---|---|");
    for (const article of report.devto) {
      lines.push(
        `| [${article.title}](${article.url}) | ${String(article.published).slice(0, 10)} | ${article.reactions} | ${article.comments} |`
      );
    }
    lines.push("");
    for (const article of report.devto) {
      if (!article.commentsList || article.commentsList.length === 0) continue;
      lines.push(`**Comments on “${article.title}”**`, "");
      for (const comment of article.commentsList) {
        lines.push(`- ${comment.author} (${String(comment.created).slice(0, 10)}): ${comment.body}`);
      }
      lines.push("");
    }
  }

  lines.push("## Reddit", "");  lines.push("| post | status | score | comments |");
  lines.push("|---|---|---|---|");
  for (const post of report.reddit) {
    lines.push(`| ${post.label} | ${post.status} | ${post.score ?? "?"} | ${post.comments ?? "?"} |`);
  }
  for (const post of report.reddit) {
    if (post.latest?.length) {
      lines.push("", `**${post.label} — newest comments**`);
      for (const comment of post.latest) {
        lines.push(`- u/${comment.author}: ${comment.body}`);
      }
    }
  }
  lines.push("");

  lines.push("## Upstream", "");
  for (const item of report.upstream) {
    const extra = item.mergedAt ? ` (merged ${item.mergedAt})` : "";
    lines.push(`- ${item.label}: **${item.state}**${extra}`);
  }
  lines.push("");

  lines.push("## Manual checks (daily, needs the browser)", "");
  lines.push(`- X thread: ${X_THREAD} — likes / follows / replies. Notifications are the source; the thread view only counts our own posts.`);
  lines.push(`- Reddit comments as the author: ${REDDIT_POSTS.map((p) => p.url).join(" · ")}`);
  lines.push("- r/SideProject modmail: still filtered unless a moderator replies; do not repost.");
  lines.push("- HN / V2EX: not posted. The body must be written and submitted by the owner; do not generate it.");
  lines.push("");

  return lines.join("\n");
}

const report = {
  localVersion: localVersion(),
  traffic: traffic(),
  releases: releases(),
  upstream: upstream(),
  devto: await devto(),
  reddit: []
};
for (const post of REDDIT_POSTS) report.reddit.push(await redditPost(post));

if (process.argv.includes("--json")) {
  const t = report.traffic || {};
  const repo = t.repo || {};
  const unreachable = [];
  if (t.error) unreachable.push("traffic");
  if (report.releases?.error) unreachable.push("releases");
  if (report.devto?.error) unreachable.push("devto");
  for (const post of report.reddit || []) {
    if (String(post.status || "").startsWith("unreachable")) unreachable.push(`reddit:${post.label}`);
  }
  for (const item of report.upstream || []) {
    if (item.state === "unreachable") unreachable.push(`upstream:${item.label}`);
  }
  const snapshot = {
    schema: "promotion-snapshot/v1",
    project_id: "ai-terminal-pet",
    generated_at: new Date().toISOString(),
    release: (() => {
      const latestTag = Array.isArray(report.releases)
        ? report.releases.map((r) => r.tag).find((tag) => /^pet-v\d+\.\d+\.\d+$/.test(tag))
        : null;
      return {
        version: latestTag ? latestTag.replace(/^pet-v/, "") : report.localVersion,
        local_version: report.localVersion,
        tag: latestTag || null,
        version_source: latestTag ? "release_tag" : "local_version"
      };
    })(),
    github: t.error ? {} : {
      stars: repo.stars,
      forks: repo.forks,
      watchers: repo.subscribers,
      open_issues: repo.issues,
      views: { count: repo.views?.count, uniques: repo.views?.uniques, window_days: 14 },
      clones: { count: repo.clones?.count, uniques: repo.clones?.uniques, ci: repo.clones?.ciClones, window_days: 14 },
      referrers: repo.referrers || []
    },
    npm: Object.entries(t.npm || {}).map(([name, n]) => {
      const share = latestShare(n);
      return {
        package: name,
        published: publishedVersion(name),
        week: n.week?.total ?? null,
        newest: share?.version ?? null,
        share: share ? Number(share.share.toFixed(4)) : null,
        window: "last-week"
      };
    }),
    catalogue: t.catalog
      ? [{ name: t.catalog.name, downloads: t.catalog.downloads, rank: `${t.catalog.rank}/${t.catalog.entries}`, median: t.catalog.median }]
      : [],
    releases: Array.isArray(report.releases) ? report.releases : [],
    social: {
      devto: Array.isArray(report.devto) ? report.devto : [],
      reddit: report.reddit || []
    },
    upstream: report.upstream || [],
    facts: {
      dsh_plugin_version: publishedVersion("dsh-ai-terminal-pet"),
      opencode_plugin_version: publishedVersion("opencode-ai-terminal-pet")
    },
    unreachable,
    source: { command: "node tools/daily-reach.mjs --json" }
  };
  process.stdout.write(JSON.stringify(snapshot, null, 2));
  process.exit(0);
}

const text = markdown(report);
process.stdout.write(text);

if (!stdoutOnly) {
  const dir = path.join(repoDir, "docs", "reach");
  mkdirSync(dir, { recursive: true });
  const day = new Date().toISOString().slice(0, 10);
  const daily = path.join(dir, `${day}.md`);
  const previous = existsSync(daily) ? readFileSync(daily, "utf8") : "";
  const separator = previous ? "\n\n---\n\n" : "";
  writeFileSync(daily, previous + separator + text, "utf8");

  // The scheduled run must not erase the manual half (X/Reddit/dev.to notes)
  // that yesterday's browser check appended to latest.md.
  const latestPath = path.join(dir, "latest.md");
  const previousLatest = existsSync(latestPath) ? readFileSync(latestPath, "utf8") : "";
  const manualAt = previousLatest.indexOf("\n## Manual check");
  const manual = manualAt >= 0 ? previousLatest.slice(manualAt) : "";
  writeFileSync(latestPath, text + manual, "utf8");
}
