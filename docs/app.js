// ===== i18n =====
var i18n = {
  en: {
    hero_l1: "AI writes code on vibes.",
    hero_l2: "This tool makes it think first.",
    hero_sub: "Go 1.25+ -- single binary -- zero dependencies",
    problem_p1: "AI agents jump straight to code. They skip requirements, write tests as afterthought, hallucinate edge cases. The result works -- until it doesn't.",
    problem_p2: "PTSD enforces a pipeline: think, then prove, then build. Every feature earns its way into the codebase.",
    step1_title: "Requirements",
    step1_desc: "Define what you're building before anyone writes a line of code. Acceptance criteria, edge cases, non-goals. The AI reads this as its contract -- not a suggestion, a constraint.",
    step2_title: "Golden Data",
    step2_desc: "Prepare real examples of what the code will handle. Not test_user or foo@bar.com -- actual records, actual edge cases. This grounds everything that follows in reality.",
    step2_note: "v2: optional. Only required in full pipeline profile.",
    step3_title: "Behavior Scenarios",
    step3_desc: "Given/When/Then specifications. Each scenario is a concrete claim: given this input, when this happens, then this is the result. Machines can verify claims. They can't verify vibes.",
    step3_note: "v2: optional for lite profile. Write tests directly from PRD.",
    step4_title: "Tests Before Code",
    step4_desc: "Write tests that fail. Red lights. They define what done means before implementation starts. No mocks for internal code -- real files, real I/O, real behavior.",
    step5_title: "Implementation",
    step5_desc: "Now write code. Only to make failing tests pass. Nothing speculative, nothing extra. The pipeline already ensured you know what to build and how to prove it works.",
    v1_title: "v1.x",
    v1_l1: "One pipeline for all features.",
    v1_l2: "Five mandatory stages. No exceptions.",
    v1_l3: "Claude Code only.",
    v1_l4: "Go-centric test detection.",
    v1_l5: "Context injected twice per message.",
    v2_title: "v2.0",
    v2_l1: "Three pipeline profiles -- full, standard, lite.",
    v2_l2: "Choose per feature. Skip what you don't need.",
    v2_l3: "Claude Code, OpenCode, Cursor -- any tool.",
    v2_l4: "Go, TypeScript, Python, Rust, Ruby, Java, C# -- any language.",
    v2_l5: "Context injected once per session. 40-60% fewer tokens.",
    versions_title: "v1 -> v2",
    start_title: "Quick Start"
  },
  ru: {
    hero_l1: "AI пишет код наугад.",
    hero_l2: "Этот инструмент заставляет его думать.",
    hero_sub: "Go 1.25+ -- один бинарник -- без зависимостей",
    problem_p1: "AI-агенты прыгают к коду напрямую. Пропускают требования, пишут тесты задним числом, галлюцинируют edge cases. Результат работает -- пока не перестаёт.",
    problem_p2: "PTSD обеспечивает пайплайн: сначала думай, потом докажи, потом строй. Каждая фича заслуживает своё место в кодовой базе.",
    step1_title: "Требования",
    step1_desc: "Определи что строишь до того, как кто-то напишет строчку кода. Критерии приёмки, крайние случаи, что НЕ делаем. AI читает это как контракт -- не рекомендацию, а ограничение.",
    step2_title: "Эталонные данные",
    step2_desc: "Подготовь реальные примеры того, с чем будет работать код. Не test_user и не foo@bar.com -- настоящие записи, настоящие крайние случаи. Это заземляет всё что идёт дальше.",
    step2_note: "v2: опционально. Только для профиля full.",
    step3_title: "Сценарии поведения",
    step3_desc: "Спецификации Given/When/Then. Каждый сценарий -- конкретное утверждение: при таких входных, когда происходит это, результат такой. Машины умеют проверять утверждения. Проверять вайбы -- нет.",
    step3_note: "v2: опционально для профиля lite. Тесты пишутся сразу из требований.",
    step4_title: "Тесты до кода",
    step4_desc: "Напиши тесты которые падают. Красные. Они определяют что значит 'готово' ещё до начала реализации. Никаких моков для внутреннего кода -- реальные файлы, реальный ввод-вывод, реальное поведение.",
    step5_title: "Реализация",
    step5_desc: "Теперь пиши код. Только чтобы тесты стали зелёными. Ничего спекулятивного, ничего лишнего. Пайплайн уже обеспечил: ты знаешь что строить и как доказать что оно работает.",
    v1_title: "v1.x",
    v1_l1: "Один пайплайн для всех фич.",
    v1_l2: "Пять обязательных стадий. Без исключений.",
    v1_l3: "Только Claude Code.",
    v1_l4: "Определение тестов только для Go.",
    v1_l5: "Контекст инжектится дважды за сообщение.",
    v2_title: "v2.0",
    v2_l1: "Три профиля пайплайна -- full, standard, lite.",
    v2_l2: "Выбирай для каждой фичи. Пропускай лишнее.",
    v2_l3: "Claude Code, OpenCode, Cursor -- любой инструмент.",
    v2_l4: "Go, TypeScript, Python, Rust, Ruby, Java, C# -- любой язык.",
    v2_l5: "Контекст инжектится раз за сессию. На 40-60% меньше токенов.",
    versions_title: "v1 -> v2",
    start_title: "Быстрый старт"
  },
  zh: {
    hero_l1: "AI写代码全凭直觉。",
    hero_l2: "这个工具让它先想清楚。",
    hero_sub: "Go 1.25+ -- 单文件 -- 零依赖",
    problem_p1: "AI助手直接跳到写代码。跳过需求分析，事后补测试，凭空编造边界情况。代码能跑\u2014\u2014直到跑不动。",
    problem_p2: "PTSD强制执行开发管线：先思考，再验证，最后构建。每个功能都要凭实力进入代码库。",
    step1_title: "需求定义",
    step1_desc: "在写任何一行代码之前，先定义要构建什么。验收标准、边界情况、明确的非目标。AI把这份文档当作约束条件来执行\u2014\u2014不是建议，是合约。",
    step2_title: "种子数据",
    step2_desc: "准备代码将要处理的真实样例。不是test_user和foo@bar.com\u2014\u2014而是真实的记录、真实的边界数据。后续所有环节都以此为基础。",
    step2_note: "v2：可选。仅在完整管线中必需。",
    step3_title: "行为场景",
    step3_desc: "用Given/When/Then编写规格说明。每个场景都是一个具体的断言：给定这个输入，当发生这件事，结果是这样。机器能验证断言，但验证不了直觉。",
    step3_note: "v2：精简管线可跳过。直接从需求文档编写测试。",
    step4_title: "先写测试",
    step4_desc: "写出会失败的测试。亮红灯。在动手实现之前，就定义好什么叫'完成'。不用模拟对象替代内部代码\u2014\u2014用真实文件、真实读写、真实行为。",
    step5_title: "实现代码",
    step5_desc: "现在写代码。只为让失败的测试变绿。不写投机代码，不加多余功能。管线已经确保你清楚要构建什么、以及如何证明它能工作。",
    v1_title: "v1.x",
    v1_l1: "所有功能共用一条管线。",
    v1_l2: "五个阶段全部必须完成，无例外。",
    v1_l3: "仅支持Claude Code。",
    v1_l4: "测试检测以Go语言为中心。",
    v1_l5: "每条消息注入两次上下文。",
    v2_title: "v2.0",
    v2_l1: "三种管线配置\u2014\u2014完整、标准、精简。",
    v2_l2: "按功能选择配置，跳过不需要的阶段。",
    v2_l3: "Claude Code、OpenCode、Cursor\u2014\u2014支持任意工具。",
    v2_l4: "Go、TypeScript、Python、Rust、Ruby、Java、C#\u2014\u2014支持任意语言。",
    v2_l5: "每次会话仅注入一次上下文。节省40-60%令牌消耗。",
    versions_title: "v1 -> v2",
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
