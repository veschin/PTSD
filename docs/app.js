// ===== i18n =====
const i18n = {
  en: {
    nav_overview: "Overview",
    nav_changelog: "Changelog",
    nav_quickstart: "Quick Start",
    hero_title: "PTSD",
    hero_tagline: "Structured AI Development Pipeline",
    hero_desc: "CLI tool enforcing PRD \u2192 Tests \u2192 Implementation for AI-driven projects. No skipping. No orphans. No vibes.",
    hero_badge1: "Zero deps",
    hero_badge2: "Single binary",
    hero_badge3: "22 commands",
    how_title: "How It Works",
    how_step1_title: "Context",
    how_step1_desc: "Session starts \u2014 pipeline state injected. AI sees what to do next.",
    how_step2_title: "Gate",
    how_step2_desc: "Every file write is checked. Wrong stage? Blocked.",
    how_step3_title: "Track",
    how_step3_desc: "Artifacts created \u2014 stage advances automatically.",
    how_step4_title: "Validate",
    how_step4_desc: "Commit hook runs full check. Nothing slips through.",
    profiles_title: "Pipeline Profiles",
    profiles_desc: "Not every feature needs five stages. Choose what fits.",
    profile_full_name: "full",
    profile_full_desc: "Data-heavy features. Golden seed data grounds BDD scenarios.",
    profile_std_name: "standard",
    profile_std_desc: "Default. Most features. BDD from PRD, skip seed ceremony.",
    profile_lite_name: "lite",
    profile_lite_desc: "Simple utilities. Tests directly from PRD. Ship fast.",
    features_title: "What\u2019s New in v2",
    feat1_title: "Pipeline Profiles",
    feat1_desc: "Three profiles per feature. Full, standard, lite. Skip what you don\u2019t need.",
    feat2_title: "Polyglot",
    feat2_desc: "Go, TypeScript, Python, Rust, Ruby, Java, C#. Auto-detected.",
    feat3_title: "Tool Adapters",
    feat3_desc: "Claude Code, OpenCode, Cursor. One pipeline, any tool.",
    feat4_title: "Smart Adopt",
    feat4_desc: "Bootstrap existing projects. Features auto-discovered from tests.",
    changelog_title: "Changelog",
    changelog_v2: "v2.0",
    changelog_v1: "v1.x",
    changelog_v1_desc: "Original release. Fixed 5-stage pipeline (PRD \u2192 Seed \u2192 BDD \u2192 Tests \u2192 Impl), Claude Code integration, Go-centric test detection. Validated across 4 benchmark rounds with zero bypass attempts by R4.",
    quickstart_title: "Quick Start",
    footer_tagline: "Made for AI-driven development",
    cl_th_aspect: "Aspect",
    cl_r1_aspect: "Pipeline", cl_r1_v1: "5 stages, mandatory", cl_r1_v2: "3 profiles: full/standard/lite",
    cl_r2_aspect: "Tools", cl_r2_v1: "Claude Code only", cl_r2_v2: "Claude + OpenCode + generic",
    cl_r3_aspect: "Languages", cl_r3_v1: "Go-first", cl_r3_v2: "Go, TS, JS, Python, Rust, Ruby, Java, C#",
    cl_r4_aspect: "Context hook", cl_r4_v1: "2x per message", cl_r4_v2: "1x per session",
    cl_r5_aspect: "adopt", cl_r5_v1: "BDD tags only", cl_r5_v2: "Test files + deferred + runner detect",
    cl_r6_aspect: "test map", cl_r6_v1: "BDD file required", cl_r6_v2: "BDD or --feature (lite)",
    cl_r7_aspect: "Migration", cl_r7_v1: "none", cl_r7_v2: "ptsd migrate"
  },
  ru: {
    nav_overview: "\u041e\u0431\u0437\u043e\u0440",
    nav_changelog: "\u0418\u0437\u043c\u0435\u043d\u0435\u043d\u0438\u044f",
    nav_quickstart: "\u041d\u0430\u0447\u0430\u043b\u043e",
    hero_title: "PTSD",
    hero_tagline: "\u0421\u0442\u0440\u0443\u043a\u0442\u0443\u0440\u0438\u0440\u043e\u0432\u0430\u043d\u043d\u044b\u0439 \u043f\u0430\u0439\u043f\u043b\u0430\u0439\u043d AI-\u0440\u0430\u0437\u0440\u0430\u0431\u043e\u0442\u043a\u0438",
    hero_desc: "CLI-\u0438\u043d\u0441\u0442\u0440\u0443\u043c\u0435\u043d\u0442 \u0434\u043b\u044f \u043f\u0430\u0439\u043f\u043b\u0430\u0439\u043d\u0430 PRD \u2192 Tests \u2192 Implementation \u0432 AI-\u043f\u0440\u043e\u0435\u043a\u0442\u0430\u0445. \u0411\u0435\u0437 \u043f\u0440\u043e\u043f\u0443\u0441\u043a\u043e\u0432. \u0411\u0435\u0437 \u043c\u0443\u0441\u043e\u0440\u0430. \u0411\u0435\u0437 \u0432\u0430\u0439\u0431-\u043a\u043e\u0434\u0438\u043d\u0433\u0430.",
    hero_badge1: "\u0411\u0435\u0437 \u0437\u0430\u0432\u0438\u0441\u0438\u043c\u043e\u0441\u0442\u0435\u0439",
    hero_badge2: "\u041e\u0434\u0438\u043d \u0431\u0438\u043d\u0430\u0440\u043d\u0438\u043a",
    hero_badge3: "22 \u043a\u043e\u043c\u0430\u043d\u0434\u044b",
    how_title: "\u041a\u0430\u043a \u044d\u0442\u043e \u0440\u0430\u0431\u043e\u0442\u0430\u0435\u0442",
    how_step1_title: "\u041a\u043e\u043d\u0442\u0435\u043a\u0441\u0442",
    how_step1_desc: "\u0421\u0435\u0441\u0441\u0438\u044f \u043d\u0430\u0447\u0438\u043d\u0430\u0435\u0442\u0441\u044f \u2014 \u0441\u043e\u0441\u0442\u043e\u044f\u043d\u0438\u0435 \u043f\u0430\u0439\u043f\u043b\u0430\u0439\u043d\u0430 \u0438\u043d\u0436\u0435\u043a\u0442\u0438\u0442\u0441\u044f. AI \u0432\u0438\u0434\u0438\u0442 \u0447\u0442\u043e \u0434\u0435\u043b\u0430\u0442\u044c.",
    how_step2_title: "\u0413\u0435\u0439\u0442",
    how_step2_desc: "\u041a\u0430\u0436\u0434\u0430\u044f \u0437\u0430\u043f\u0438\u0441\u044c \u0444\u0430\u0439\u043b\u0430 \u043f\u0440\u043e\u0432\u0435\u0440\u044f\u0435\u0442\u0441\u044f. \u041d\u0435 \u0442\u0430 \u0441\u0442\u0430\u0434\u0438\u044f? \u0417\u0430\u0431\u043b\u043e\u043a\u0438\u0440\u043e\u0432\u0430\u043d\u043e.",
    how_step3_title: "\u0422\u0440\u0435\u043a\u0438\u043d\u0433",
    how_step3_desc: "\u0410\u0440\u0442\u0435\u0444\u0430\u043a\u0442 \u0441\u043e\u0437\u0434\u0430\u043d \u2014 \u0441\u0442\u0430\u0434\u0438\u044f \u043f\u0440\u043e\u0434\u0432\u0438\u0433\u0430\u0435\u0442\u0441\u044f \u0430\u0432\u0442\u043e\u043c\u0430\u0442\u0438\u0447\u0435\u0441\u043a\u0438.",
    how_step4_title: "\u0412\u0430\u043b\u0438\u0434\u0430\u0446\u0438\u044f",
    how_step4_desc: "\u0425\u0443\u043a \u043a\u043e\u043c\u043c\u0438\u0442\u0430 \u0437\u0430\u043f\u0443\u0441\u043a\u0430\u0435\u0442 \u043f\u043e\u043b\u043d\u0443\u044e \u043f\u0440\u043e\u0432\u0435\u0440\u043a\u0443. \u041d\u0438\u0447\u0442\u043e \u043d\u0435 \u043f\u0440\u043e\u0441\u043a\u043e\u0447\u0438\u0442.",
    profiles_title: "\u041f\u0440\u043e\u0444\u0438\u043b\u0438 \u043f\u0430\u0439\u043f\u043b\u0430\u0439\u043d\u0430",
    profiles_desc: "\u041d\u0435 \u043a\u0430\u0436\u0434\u0430\u044f \u0444\u0438\u0447\u0430 \u0442\u0440\u0435\u0431\u0443\u0435\u0442 \u043f\u044f\u0442\u0438 \u0441\u0442\u0430\u0434\u0438\u0439. \u0412\u044b\u0431\u0438\u0440\u0430\u0439 \u0447\u0442\u043e \u043f\u043e\u0434\u0445\u043e\u0434\u0438\u0442.",
    profile_full_name: "full",
    profile_full_desc: "\u0421\u043b\u043e\u0436\u043d\u044b\u0435 \u0444\u0438\u0447\u0438 \u0441 \u0434\u0430\u043d\u043d\u044b\u043c\u0438. \u0421\u0438\u0434-\u0434\u0430\u043d\u043d\u044b\u0435 \u043e\u0431\u043e\u0441\u043d\u043e\u0432\u044b\u0432\u0430\u044e\u0442 BDD-\u0441\u0446\u0435\u043d\u0430\u0440\u0438\u0438.",
    profile_std_name: "standard",
    profile_std_desc: "\u041f\u043e \u0443\u043c\u043e\u043b\u0447\u0430\u043d\u0438\u044e. \u0411\u043e\u043b\u044c\u0448\u0438\u043d\u0441\u0442\u0432\u043e \u0444\u0438\u0447. BDD \u0438\u0437 PRD, \u0431\u0435\u0437 \u0441\u0438\u0434-\u0446\u0435\u0440\u0435\u043c\u043e\u043d\u0438\u0439.",
    profile_lite_name: "lite",
    profile_lite_desc: "\u041f\u0440\u043e\u0441\u0442\u044b\u0435 \u0443\u0442\u0438\u043b\u0438\u0442\u044b. \u0422\u0435\u0441\u0442\u044b \u043f\u0440\u044f\u043c\u043e \u0438\u0437 PRD. \u0411\u044b\u0441\u0442\u0440\u0430\u044f \u043f\u043e\u0441\u0442\u0430\u0432\u043a\u0430.",
    features_title: "\u0427\u0442\u043e \u043d\u043e\u0432\u043e\u0433\u043e \u0432 v2",
    feat1_title: "\u041f\u0440\u043e\u0444\u0438\u043b\u0438 \u043f\u0430\u0439\u043f\u043b\u0430\u0439\u043d\u0430",
    feat1_desc: "\u0422\u0440\u0438 \u043f\u0440\u043e\u0444\u0438\u043b\u044f \u043d\u0430 \u0444\u0438\u0447\u0443. Full, standard, lite. \u041f\u0440\u043e\u043f\u0443\u0441\u043a\u0430\u0439 \u043b\u0438\u0448\u043d\u0435\u0435.",
    feat2_title: "\u041f\u043e\u043b\u0438\u0433\u043b\u043e\u0442",
    feat2_desc: "Go, TypeScript, Python, Rust, Ruby, Java, C#. \u0410\u0432\u0442\u043e\u043e\u043f\u0440\u0435\u0434\u0435\u043b\u0435\u043d\u0438\u0435.",
    feat3_title: "\u0410\u0434\u0430\u043f\u0442\u0435\u0440\u044b \u0438\u043d\u0441\u0442\u0440\u0443\u043c\u0435\u043d\u0442\u043e\u0432",
    feat3_desc: "Claude Code, OpenCode, Cursor. \u041e\u0434\u0438\u043d \u043f\u0430\u0439\u043f\u043b\u0430\u0439\u043d, \u043b\u044e\u0431\u043e\u0439 \u0438\u043d\u0441\u0442\u0440\u0443\u043c\u0435\u043d\u0442.",
    feat4_title: "\u0423\u043c\u043d\u044b\u0439 Adopt",
    feat4_desc: "\u041f\u043e\u0434\u043a\u043b\u044e\u0447\u0435\u043d\u0438\u0435 \u043a \u0441\u0443\u0449\u0435\u0441\u0442\u0432\u0443\u044e\u0449\u0438\u043c \u043f\u0440\u043e\u0435\u043a\u0442\u0430\u043c. \u0424\u0438\u0447\u0438 \u043d\u0430\u0445\u043e\u0434\u044f\u0442\u0441\u044f \u0438\u0437 \u0442\u0435\u0441\u0442\u043e\u0432.",
    changelog_title: "\u0418\u0437\u043c\u0435\u043d\u0435\u043d\u0438\u044f",
    changelog_v2: "v2.0",
    changelog_v1: "v1.x",
    changelog_v1_desc: "\u041f\u0435\u0440\u0432\u044b\u0439 \u0440\u0435\u043b\u0438\u0437. \u0416\u0451\u0441\u0442\u043a\u0438\u0439 5-\u0441\u0442\u0430\u0434\u0438\u0439\u043d\u044b\u0439 \u043f\u0430\u0439\u043f\u043b\u0430\u0439\u043d (PRD \u2192 Seed \u2192 BDD \u2192 Tests \u2192 Impl), \u0438\u043d\u0442\u0435\u0433\u0440\u0430\u0446\u0438\u044f \u0441 Claude Code, Go-\u0446\u0435\u043d\u0442\u0440\u0438\u0447\u043d\u043e\u0435 \u043e\u043f\u0440\u0435\u0434\u0435\u043b\u0435\u043d\u0438\u0435 \u0442\u0435\u0441\u0442\u043e\u0432. \u041f\u0440\u043e\u0432\u0435\u0440\u0435\u043d \u0432 4 \u0440\u0430\u0443\u043d\u0434\u0430\u0445 \u0431\u0435\u043d\u0447\u043c\u0430\u0440\u043a\u043e\u0432, \u043d\u043e\u043b\u044c \u043e\u0431\u0445\u043e\u0434\u043e\u0432 \u043a R4.",
    quickstart_title: "\u0411\u044b\u0441\u0442\u0440\u044b\u0439 \u0441\u0442\u0430\u0440\u0442",
    footer_tagline: "\u0414\u043b\u044f AI-\u0443\u043f\u0440\u0430\u0432\u043b\u044f\u0435\u043c\u043e\u0439 \u0440\u0430\u0437\u0440\u0430\u0431\u043e\u0442\u043a\u0438",
    cl_th_aspect: "\u0410\u0441\u043f\u0435\u043a\u0442",
    cl_r1_aspect: "\u041f\u0430\u0439\u043f\u043b\u0430\u0439\u043d", cl_r1_v1: "5 \u0441\u0442\u0430\u0434\u0438\u0439, \u043e\u0431\u044f\u0437\u0430\u0442\u0435\u043b\u044c\u043d\u043e", cl_r1_v2: "3 \u043f\u0440\u043e\u0444\u0438\u043b\u044f: full/standard/lite",
    cl_r2_aspect: "\u0418\u043d\u0441\u0442\u0440\u0443\u043c\u0435\u043d\u0442\u044b", cl_r2_v1: "\u0422\u043e\u043b\u044c\u043a\u043e Claude Code", cl_r2_v2: "Claude + OpenCode + generic",
    cl_r3_aspect: "\u042f\u0437\u044b\u043a\u0438", cl_r3_v1: "Go-first", cl_r3_v2: "Go, TS, JS, Python, Rust, Ruby, Java, C#",
    cl_r4_aspect: "\u041a\u043e\u043d\u0442\u0435\u043a\u0441\u0442-\u0445\u0443\u043a", cl_r4_v1: "2\u00d7 \u0437\u0430 \u0441\u043e\u043e\u0431\u0449\u0435\u043d\u0438\u0435", cl_r4_v2: "1\u00d7 \u0437\u0430 \u0441\u0435\u0441\u0441\u0438\u044e",
    cl_r5_aspect: "adopt", cl_r5_v1: "\u0422\u043e\u043b\u044c\u043a\u043e BDD-\u0442\u0435\u0433\u0438", cl_r5_v2: "\u0422\u0435\u0441\u0442-\u0444\u0430\u0439\u043b\u044b + deferred + detect runner",
    cl_r6_aspect: "test map", cl_r6_v1: "\u041d\u0443\u0436\u0435\u043d BDD-\u0444\u0430\u0439\u043b", cl_r6_v2: "BDD \u0438\u043b\u0438 --feature (lite)",
    cl_r7_aspect: "\u041c\u0438\u0433\u0440\u0430\u0446\u0438\u044f", cl_r7_v1: "\u041d\u0435\u0442", cl_r7_v2: "ptsd migrate"
  },
  zh: {
    nav_overview: "\u6982\u89c8",
    nav_changelog: "\u66f4\u65b0\u65e5\u5fd7",
    nav_quickstart: "\u5feb\u901f\u5f00\u59cb",
    hero_title: "PTSD",
    hero_tagline: "\u7ed3\u6784\u5316AI\u5f00\u53d1\u7ba1\u7ebf",
    hero_desc: "\u5f3a\u5236\u6267\u884c PRD \u2192 Tests \u2192 Implementation \u7ba1\u7ebf\u7684CLI\u5de5\u5177\u3002\u4e0d\u8df3\u8fc7\u9636\u6bb5\u3002\u4e0d\u9057\u7559\u5b64\u7acb\u6587\u4ef6\u3002\u4e0d\u9760\u76f4\u89c9\u53d1\u5e03\u3002",
    hero_badge1: "\u96f6\u4f9d\u8d56",
    hero_badge2: "\u5355\u4e8c\u8fdb\u5236\u6587\u4ef6",
    hero_badge3: "22\u4e2a\u547d\u4ee4",
    how_title: "\u5de5\u4f5c\u539f\u7406",
    how_step1_title: "\u4e0a\u4e0b\u6587",
    how_step1_desc: "\u4f1a\u8bdd\u5f00\u59cb\u65f6\u6ce8\u5165\u7ba1\u7ebf\u72b6\u6001\u3002AI\u6e05\u695a\u4e0b\u4e00\u6b65\u505a\u4ec0\u4e48\u3002",
    how_step2_title: "\u95e8\u63a7",
    how_step2_desc: "\u6bcf\u6b21\u6587\u4ef6\u5199\u5165\u90fd\u7ecf\u8fc7\u68c0\u67e5\u3002\u9636\u6bb5\u4e0d\u5bf9\uff1f\u963b\u6b62\u3002",
    how_step3_title: "\u8ffd\u8e2a",
    how_step3_desc: "\u5de5\u4ef6\u521b\u5efa\u540e\uff0c\u9636\u6bb5\u81ea\u52a8\u63a8\u8fdb\u3002",
    how_step4_title: "\u9a8c\u8bc1",
    how_step4_desc: "\u63d0\u4ea4\u94a9\u5b50\u8fd0\u884c\u5b8c\u6574\u68c0\u67e5\u3002\u4efb\u4f55\u8fdd\u89c4\u90fd\u65e0\u6cd5\u901a\u8fc7\u3002",
    profiles_title: "\u7ba1\u7ebf\u914d\u7f6e",
    profiles_desc: "\u5e76\u975e\u6bcf\u4e2a\u529f\u80fd\u90fd\u9700\u8981\u4e94\u4e2a\u9636\u6bb5\u3002\u9009\u62e9\u9002\u5408\u7684\u914d\u7f6e\u3002",
    profile_full_name: "full",
    profile_full_desc: "\u6570\u636e\u5bc6\u96c6\u578b\u529f\u80fd\u3002\u79cd\u5b50\u6570\u636e\u652f\u6491BDD\u573a\u666f\u3002",
    profile_std_name: "standard",
    profile_std_desc: "\u9ed8\u8ba4\u914d\u7f6e\u3002\u5927\u591a\u6570\u529f\u80fd\u3002\u4ecePRD\u7f16\u5199BDD\uff0c\u8df3\u8fc7\u79cd\u5b50\u73af\u8282\u3002",
    profile_lite_name: "lite",
    profile_lite_desc: "\u7b80\u5355\u5de5\u5177\u7c7b\u3002\u76f4\u63a5\u4ecePRD\u7f16\u5199\u6d4b\u8bd5\u3002\u5feb\u901f\u4ea4\u4ed8\u3002",
    features_title: "v2\u65b0\u7279\u6027",
    feat1_title: "\u7ba1\u7ebf\u914d\u7f6e",
    feat1_desc: "\u6bcf\u4e2a\u529f\u80fd\u4e09\u79cd\u914d\u7f6e\u3002Full\u3001standard\u3001lite\u3002\u8df3\u8fc7\u4e0d\u9700\u8981\u7684\u3002",
    feat2_title: "\u591a\u8bed\u8a00",
    feat2_desc: "Go\u3001TypeScript\u3001Python\u3001Rust\u3001Ruby\u3001Java\u3001C#\u3002\u81ea\u52a8\u68c0\u6d4b\u3002",
    feat3_title: "\u5de5\u5177\u9002\u914d\u5668",
    feat3_desc: "Claude Code\u3001OpenCode\u3001Cursor\u3002\u4e00\u6761\u7ba1\u7ebf\uff0c\u4efb\u4f55\u5de5\u5177\u3002",
    feat4_title: "\u667a\u80fd\u63a5\u5165",
    feat4_desc: "\u63a5\u5165\u73b0\u6709\u9879\u76ee\u3002\u4ece\u6d4b\u8bd5\u6587\u4ef6\u81ea\u52a8\u53d1\u73b0\u529f\u80fd\u3002",
    changelog_title: "\u66f4\u65b0\u65e5\u5fd7",
    changelog_v2: "v2.0",
    changelog_v1: "v1.x",
    changelog_v1_desc: "\u9996\u6b21\u53d1\u5e03\u3002\u56fa\u5b9a5\u9636\u6bb5\u7ba1\u7ebf\uff08PRD \u2192 Seed \u2192 BDD \u2192 Tests \u2192 Impl\uff09\uff0cClaude Code\u96c6\u6210\uff0cGo\u4e2d\u5fc3\u5316\u6d4b\u8bd5\u68c0\u6d4b\u3002\u7ecf\u8fc74\u8f6e\u57fa\u51c6\u6d4b\u8bd5\u9a8c\u8bc1\uff0c\u7b2c4\u8f6e\u96f6\u7ed5\u8fc7\u5c1d\u8bd5\u3002",
    quickstart_title: "\u5feb\u901f\u5f00\u59cb",
    footer_tagline: "\u4e3aAI\u9a71\u52a8\u7684\u5f00\u53d1\u800c\u751f",
    cl_th_aspect: "\u65b9\u9762",
    cl_r1_aspect: "\u7ba1\u7ebf", cl_r1_v1: "5\u9636\u6bb5\uff0c\u5f3a\u5236", cl_r1_v2: "3\u79cd\u914d\u7f6e\uff1afull/standard/lite",
    cl_r2_aspect: "\u5de5\u5177", cl_r2_v1: "\u4ec5Claude Code", cl_r2_v2: "Claude + OpenCode + generic",
    cl_r3_aspect: "\u8bed\u8a00", cl_r3_v1: "Go\u4f18\u5148", cl_r3_v2: "Go, TS, JS, Python, Rust, Ruby, Java, C#",
    cl_r4_aspect: "\u4e0a\u4e0b\u6587\u94a9\u5b50", cl_r4_v1: "\u6bcf\u6d88\u606f2\u6b21", cl_r4_v2: "\u6bcf\u4f1a\u8bdd1\u6b21",
    cl_r5_aspect: "adopt", cl_r5_v1: "\u4ec5BDD\u6807\u7b7e", cl_r5_v2: "\u6d4b\u8bd5\u6587\u4ef6 + deferred + runner\u68c0\u6d4b",
    cl_r6_aspect: "test map", cl_r6_v1: "\u9700\u8981BDD\u6587\u4ef6", cl_r6_v2: "BDD\u6216--feature\uff08lite\uff09",
    cl_r7_aspect: "\u8fc1\u79fb", cl_r7_v1: "\u65e0", cl_r7_v2: "ptsd migrate"
  }
};

// ===== Locale =====
let currentLocale = localStorage.getItem("ptsd-locale") || "en";

function applyLocale(locale) {
  currentLocale = locale;
  localStorage.setItem("ptsd-locale", locale);

  document.querySelectorAll(".locale-btn").forEach(function(btn) {
    btn.classList.toggle("active", btn.dataset.locale === locale);
  });

  document.documentElement.lang = locale === "zh" ? "zh-CN" : locale;

  var data = i18n[locale] || i18n.en;
  document.querySelectorAll("[data-i18n]").forEach(function(el) {
    var key = el.dataset.i18n;
    if (data[key] !== undefined) {
      if (el.querySelector("code")) {
        // preserve inner HTML for elements with <code>
        el.innerHTML = data[key];
      } else {
        el.textContent = data[key];
      }
    }
  });
}

function switchLocale(locale) {
  var wrapper = document.querySelector(".locale-wrapper");
  wrapper.classList.add("fading");
  setTimeout(function() {
    applyLocale(locale);
    wrapper.classList.remove("fading");
  }, 200);
}

// ===== Scroll Reveal =====
function initReveal() {
  var observer = new IntersectionObserver(function(entries) {
    entries.forEach(function(entry) {
      if (entry.isIntersecting) {
        entry.target.classList.add("visible");
      }
    });
  }, { threshold: 0.15 });

  document.querySelectorAll(".reveal").forEach(function(el) {
    observer.observe(el);
  });
}

// ===== Staggered reveal for pipeline pills =====
function initStagger() {
  document.querySelectorAll(".profile-row").forEach(function(row) {
    var pills = row.querySelectorAll(".stage-pill");
    pills.forEach(function(pill, i) {
      pill.style.transitionDelay = (i * 0.08) + "s";
    });
  });
}

// ===== Nav scroll effect =====
function initNavScroll() {
  var nav = document.getElementById("nav");
  window.addEventListener("scroll", function() {
    nav.classList.toggle("scrolled", window.scrollY > 10);
  }, { passive: true });
}

// ===== Copy button =====
function initCopy() {
  var btn = document.getElementById("copy-install");
  if (!btn) return;
  btn.addEventListener("click", function() {
    var text = "go install github.com/veschin/ptsd/cmd/ptsd@latest";
    navigator.clipboard.writeText(text).then(function() {
      btn.classList.add("copied");
      setTimeout(function() { btn.classList.remove("copied"); }, 1500);
    });
  });
}

// ===== Init =====
document.addEventListener("DOMContentLoaded", function() {
  // Lucide icons
  if (window.lucide) lucide.createIcons();

  // Locale buttons
  document.querySelectorAll(".locale-btn").forEach(function(btn) {
    btn.addEventListener("click", function() {
      switchLocale(btn.dataset.locale);
    });
  });

  // Apply saved or default locale
  applyLocale(currentLocale);

  // Animations
  initReveal();
  initStagger();
  initNavScroll();
  initCopy();
});
