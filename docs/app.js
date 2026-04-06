// ===== i18n =====
var i18n = {
  en: {
    hero_l1: "AI writes code on vibes.",
    hero_l2: "This tool makes it think first.",
    hero_sub: "Go 1.25+ -- single binary -- zero dependencies",
    hero_cta: "Quick Start \u2193",
    problem_p1: "AI agents jump straight to code. They skip requirements, write tests as afterthought, hallucinate edge cases. The result works -- until it doesn't.",
    problem_p2: "PTSD enforces a pipeline: think, then prove, then build. Every feature earns its way into the codebase.",
    step1_title: "Requirements",
    step1_desc: "Define what you're building before anyone writes a line of code. Acceptance criteria, edge cases, non-goals. The AI reads this as its contract -- not a suggestion, a constraint.",
    step2_title: "Golden Data",
    step2_desc: "Prepare real examples of what the code will handle. Not test_user or foo@bar.com -- actual records, actual edge cases. This grounds everything that follows in reality.",
    step2_note: "Optional. Only required in the full pipeline profile.",
    step3_title: "Behavior Scenarios",
    step3_desc: "Given/When/Then specifications. Each scenario is a concrete claim: given this input, when this happens, then this is the result. Machines can verify claims. They can't verify vibes.",
    step3_note: "Optional for lite profile. Write tests directly from PRD.",
    step4_title: "Tests Before Code",
    step4_desc: "Write tests that fail. Red lights. They define what done means before implementation starts. No mocks for internal code -- real files, real I/O, real behavior.",
    step5_title: "Implementation",
    step5_desc: "Now write code. Only to make failing tests pass. Nothing speculative, nothing extra. The pipeline already ensured you know what to build and how to prove it works.",
    profiles_title: "Not every feature needs five stages",
    profiles_intro: "Choose a pipeline profile per feature. Complex data-heavy work gets the full treatment. Simple utilities skip what they don't need.",
    enforce_title: "How enforcement works",
    enforce_intro: "PTSD doesn't rely on the AI following instructions voluntarily. Hooks intercept every file write and every commit.",
    enforce_s1_title: "Session Start",
    enforce_s1_desc: "Context injected once. AI sees current stage and next action per feature.",
    enforce_s2_title: "File Write",
    enforce_s2_desc: "Gate checks stage. Wrong stage? Write is blocked before it happens.",
    enforce_s3_title: "After Write",
    enforce_s3_desc: "Stage advances automatically. No manual tracking needed.",
    enforce_s4_title: "Commit",
    enforce_s4_desc: "Validate runs. Broken pipeline, bad format -- commit is blocked.",
    integ_title: "Works with your tool",
    integ_intro: "Run ptsd init -- it detects your AI tool and generates the right integration. Switch tools later with ptsd init --tool <name>.",
    integ_claude_note: "Full enforcement. Every write checked, every stage tracked.",
    integ_opencode_note: "Plugin-based enforcement. Same pipeline, different tool.",
    integ_generic_note: "Git hooks enforce at commit. AI follows instructions voluntarily.",
    skills_title: "13 specialized instructions",
    skills_intro: "Each pipeline stage has a skill -- a focused instruction set that tells the AI exactly what to produce, what to check, and what mistakes to avoid. The AI doesn't guess how to write BDD scenarios -- it follows a concrete checklist.",
    versions_title: "What changed",
    v1_title: "v1.x",
    v2_title: "v2.0",
    ct_pipeline: "Pipeline",
    ct_pipeline_v1: "5 stages, all mandatory",
    ct_pipeline_v2: "3 profiles: full, standard, lite",
    ct_tools: "Tools",
    ct_tools_v1: "Claude Code",
    ct_tools_v2: "Claude Code + OpenCode + any",
    ct_langs: "Languages",
    ct_langs_v1: "Go",
    ct_langs_v2: "Go, TS, JS, Python, Rust, Ruby, Java, C#",
    ct_context: "Context cost",
    ct_context_v1: "2x per message",
    ct_context_v2: "1x per session (-60%)",
    ct_existing: "Existing projects",
    ct_existing_v1: "Manual setup",
    ct_existing_v2: "ptsd adopt --auto-discover",
    ct_testmap: "Test mapping",
    ct_testmap_v1: "BDD file required",
    ct_testmap_v2: "Direct or BDD",
    ct_migrate: "Migration",
    ct_migrate_v1: "--",
    ct_migrate_v2: "ptsd migrate",
    start_title: "Quick Start"
  },
  ru: {
    hero_l1: "AI пишет код наугад.",
    hero_l2: "Этот инструмент заставляет его думать.",
    hero_sub: "Go 1.25+ -- один бинарник -- без зависимостей",
    hero_cta: "Быстрый старт \u2193",
    problem_p1: "AI-агенты прыгают к коду напрямую. Пропускают требования, пишут тесты задним числом, галлюцинируют edge cases. Результат работает -- пока не перестаёт.",
    problem_p2: "PTSD обеспечивает пайплайн: сначала думай, потом докажи, потом строй. Каждая фича заслуживает своё место в кодовой базе.",
    step1_title: "Требования",
    step1_desc: "Определи что строишь до того, как кто-то напишет строчку кода. Критерии приёмки, крайние случаи, что НЕ делаем. AI читает это как контракт -- не рекомендацию, а ограничение.",
    step2_title: "Эталонные данные",
    step2_desc: "Подготовь реальные примеры того, с чем будет работать код. Не test_user и не foo@bar.com -- настоящие записи, настоящие крайние случаи. Это заземляет всё что идёт дальше.",
    step2_note: "Опционально. Только для профиля full.",
    step3_title: "Сценарии поведения",
    step3_desc: "Спецификации Given/When/Then. Каждый сценарий -- конкретное утверждение: при таких входных, когда происходит это, результат такой. Машины умеют проверять утверждения. Проверять вайбы -- нет.",
    step3_note: "Опционально для профиля lite. Тесты пишутся сразу из требований.",
    step4_title: "Тесты до кода",
    step4_desc: "Напиши тесты которые падают. Красные. Они определяют что значит 'готово' ещё до начала реализации. Никаких моков для внутреннего кода -- реальные файлы, реальный ввод-вывод, реальное поведение.",
    step5_title: "Реализация",
    step5_desc: "Теперь пиши код. Только чтобы тесты стали зелёными. Ничего спекулятивного, ничего лишнего. Пайплайн уже обеспечил: ты знаешь что строить и как доказать что оно работает.",
    profiles_title: "Не каждой фиче нужны все пять стадий",
    profiles_intro: "Выбирай профиль пайплайна для каждой фичи. Сложная работа с данными -- полный цикл. Простые утилиты -- пропускают лишнее.",
    enforce_title: "Как работает контроль",
    enforce_intro: "PTSD не рассчитывает на то, что AI добровольно следует инструкциям. Хуки перехватывают каждую запись файла и каждый коммит.",
    enforce_s1_title: "Старт сессии",
    enforce_s1_desc: "Контекст инжектируется один раз. AI видит текущую стадию и следующее действие для каждой фичи.",
    enforce_s2_title: "Запись файла",
    enforce_s2_desc: "Гейт проверяет стадию. Не та стадия? Запись блокируется до того, как произойдёт.",
    enforce_s3_title: "После записи",
    enforce_s3_desc: "Стадия продвигается автоматически. Ручное отслеживание не нужно.",
    enforce_s4_title: "Коммит",
    enforce_s4_desc: "Запускается валидация. Сломанный пайплайн, плохой формат -- коммит блокируется.",
    integ_title: "Работает с твоим инструментом",
    integ_intro: "Запусти ptsd init -- он определит твой AI-инструмент и сгенерирует нужную интеграцию. Смени инструмент позже через ptsd init --tool <name>.",
    integ_claude_note: "Полный контроль. Каждая запись проверяется, каждая стадия отслеживается.",
    integ_opencode_note: "Контроль через плагин. Тот же пайплайн, другой инструмент.",
    integ_generic_note: "Git-хуки контролируют при коммите. AI следует инструкциям добровольно.",
    skills_title: "13 специализированных инструкций",
    skills_intro: "Для каждой стадии пайплайна есть скилл -- сфокусированный набор инструкций, который говорит AI что именно производить, что проверять и каких ошибок избегать. AI не угадывает как писать BDD-сценарии -- он следует конкретному чеклисту.",
    versions_title: "Что изменилось",
    v1_title: "v1.x",
    v2_title: "v2.0",
    ct_pipeline: "Пайплайн",
    ct_pipeline_v1: "5 стадий, все обязательные",
    ct_pipeline_v2: "3 профиля: full, standard, lite",
    ct_tools: "Инструменты",
    ct_tools_v1: "Claude Code",
    ct_tools_v2: "Claude Code + OpenCode + любой",
    ct_langs: "Языки",
    ct_langs_v1: "Go",
    ct_langs_v2: "Go, TS, JS, Python, Rust, Ruby, Java, C#",
    ct_context: "Стоимость контекста",
    ct_context_v1: "2x за сообщение",
    ct_context_v2: "1x за сессию (-60%)",
    ct_existing: "Существующие проекты",
    ct_existing_v1: "Ручная настройка",
    ct_existing_v2: "ptsd adopt --auto-discover",
    ct_testmap: "Маппинг тестов",
    ct_testmap_v1: "Нужен BDD-файл",
    ct_testmap_v2: "Прямой или через BDD",
    ct_migrate: "Миграция",
    ct_migrate_v1: "--",
    ct_migrate_v2: "ptsd migrate",
    start_title: "Быстрый старт"
  },
  zh: {
    hero_l1: "AI写代码全凭直觉。",
    hero_l2: "这个工具让它先想清楚。",
    hero_sub: "Go 1.25+ -- 单文件 -- 零依赖",
    hero_cta: "快速开始 \u2193",
    problem_p1: "AI助手直接跳到写代码。跳过需求分析，事后补测试，凭空编造边界情况。代码能跑\u2014\u2014直到跑不动。",
    problem_p2: "PTSD强制执行开发管线：先思考，再验证，最后构建。每个功能都要凭实力进入代码库。",
    step1_title: "需求定义",
    step1_desc: "在写任何一行代码之前，先定义要构建什么。验收标准、边界情况、明确的非目标。AI把这份文档当作约束条件来执行\u2014\u2014不是建议，是合约。",
    step2_title: "种子数据",
    step2_desc: "准备代码将要处理的真实样例。不是test_user和foo@bar.com\u2014\u2014而是真实的记录、真实的边界数据。后续所有环节都以此为基础。",
    step2_note: "可选。仅在完整管线中必需。",
    step3_title: "行为场景",
    step3_desc: "用Given/When/Then编写规格说明。每个场景都是一个具体的断言：给定这个输入，当发生这件事，结果是这样。机器能验证断言，但验证不了直觉。",
    step3_note: "精简管线可跳过。直接从需求文档编写测试。",
    step4_title: "先写测试",
    step4_desc: "写出会失败的测试。亮红灯。在动手实现之前，就定义好什么叫'完成'。不用模拟对象替代内部代码\u2014\u2014用真实文件、真实读写、真实行为。",
    step5_title: "实现代码",
    step5_desc: "现在写代码。只为让失败的测试变绿。不写投机代码，不加多余功能。管线已经确保你清楚要构建什么、以及如何证明它能工作。",
    profiles_title: "不是每个功能都需要五个阶段",
    profiles_intro: "按功能选择管线配置。复杂的数据密集型工作走完整流程。简单工具类跳过不需要的阶段。",
    enforce_title: "执行机制",
    enforce_intro: "PTSD不依赖AI自觉遵守指令。钩子拦截每一次文件写入和每一次提交。",
    enforce_s1_title: "会话启动",
    enforce_s1_desc: "上下文注入一次。AI看到每个功能的当前阶段和下一步操作。",
    enforce_s2_title: "文件写入",
    enforce_s2_desc: "门控检查阶段。阶段不对？写入在发生前被阻止。",
    enforce_s3_title: "写入后",
    enforce_s3_desc: "阶段自动推进。无需手动跟踪。",
    enforce_s4_title: "提交",
    enforce_s4_desc: "验证运行。管线损坏、格式不对----提交被阻止。",
    integ_title: "兼容你的工具",
    integ_intro: "运行ptsd init----它自动检测你的AI工具并生成对应集成。稍后通过ptsd init --tool <name>切换工具。",
    integ_claude_note: "完整执行。每次写入都被检查，每个阶段都被跟踪。",
    integ_opencode_note: "基于插件的执行。相同管线，不同工具。",
    integ_generic_note: "Git钩子在提交时执行。AI自愿遵循指令。",
    skills_title: "13个专项指令",
    skills_intro: "每个管线阶段都有一个技能----一套专注的指令集，精确告诉AI产出什么、检查什么、避免哪些错误。AI不用猜测如何写BDD场景----它遵循具体的检查清单。",
    versions_title: "版本对比",
    v1_title: "v1.x",
    v2_title: "v2.0",
    ct_pipeline: "管线",
    ct_pipeline_v1: "5个阶段，全部必须",
    ct_pipeline_v2: "3种配置：完整、标准、精简",
    ct_tools: "工具",
    ct_tools_v1: "Claude Code",
    ct_tools_v2: "Claude Code + OpenCode + 任意",
    ct_langs: "语言",
    ct_langs_v1: "Go",
    ct_langs_v2: "Go、TS、JS、Python、Rust、Ruby、Java、C#",
    ct_context: "上下文消耗",
    ct_context_v1: "每条消息2次",
    ct_context_v2: "每次会话1次（-60%）",
    ct_existing: "现有项目",
    ct_existing_v1: "手动配置",
    ct_existing_v2: "ptsd adopt --auto-discover",
    ct_testmap: "测试映射",
    ct_testmap_v1: "需要BDD文件",
    ct_testmap_v2: "直接映射或通过BDD",
    ct_migrate: "迁移",
    ct_migrate_v1: "--",
    ct_migrate_v2: "ptsd migrate",
    start_title: "快速开始"
  }
};

// ===== Locale =====
var currentLocale = localStorage.getItem("ptsd-locale") || "en";

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
      el.textContent = data[key];
    }
  });
}

function switchLocale(locale) {
  var wrapper = document.querySelector(".locale-wrapper");
  wrapper.classList.add("fading");
  setTimeout(function() {
    applyLocale(locale);
    wrapper.classList.remove("fading");
  }, 150);
}

// ===== Clip-path Reveal =====
function initReveal() {
  if (!("IntersectionObserver" in window)) {
    // Fallback: show everything immediately
    document.querySelectorAll(".reveal").forEach(function(el) {
      el.style.clipPath = "none";
    });
    return;
  }

  var observer = new IntersectionObserver(function(entries) {
    entries.forEach(function(entry) {
      if (entry.isIntersecting) {
        entry.target.classList.add("visible");
        observer.unobserve(entry.target);
      }
    });
  }, { threshold: 0.1 });

  document.querySelectorAll(".reveal").forEach(function(el) {
    observer.observe(el);
  });
}

// ===== Line Reveal =====
function initLineReveal() {
  if (!("IntersectionObserver" in window)) return;

  var observer = new IntersectionObserver(function(entries) {
    entries.forEach(function(entry) {
      if (entry.isIntersecting) {
        entry.target.classList.add("visible");
        observer.unobserve(entry.target);
      }
    });
  }, { threshold: 0.2 });

  document.querySelectorAll(".line-reveal").forEach(function(el) {
    // Mark as animating so CSS hides children
    el.classList.add("animating");
    observer.observe(el);
  });
}

// ===== Nav scroll effect =====
function initNavScroll() {
  var nav = document.getElementById("nav");
  if (!nav) return;
  window.addEventListener("scroll", function() {
    nav.classList.toggle("scrolled", window.scrollY > 10);
  }, { passive: true });
}

// ===== Init =====
document.addEventListener("DOMContentLoaded", function() {
  // Locale buttons (nav + footer share same class)
  document.querySelectorAll(".locale-btn").forEach(function(btn) {
    btn.addEventListener("click", function() {
      switchLocale(btn.dataset.locale);
    });
  });

  applyLocale(currentLocale);
  initReveal();
  initLineReveal();
  initNavScroll();
});
