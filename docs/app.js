// ===== Skills data -- fetched from GitHub raw =====
var SKILLS = [
  "write-prd", "write-seed", "write-bdd", "write-tests", "write-impl",
  "review-prd", "review-seed", "review-bdd", "review-tests", "review-impl",
  "create-tasks", "workflow", "adopt"
];
var RAW_BASE = "https://raw.githubusercontent.com/veschin/PTSD/main/internal/core/templates/skills/";
var skillCache = {};

function initSkills() {
  var tabs = document.getElementById("skill-tabs");
  var view = document.getElementById("skill-view");
  if (!tabs || !view) return;

  SKILLS.forEach(function(name, i) {
    var btn = document.createElement("button");
    btn.className = "skill-tab";
    btn.textContent = name;
    btn.addEventListener("click", function() { loadSkill(name); });
    tabs.appendChild(btn);
  });

  // load first skill
  loadSkill(SKILLS[0]);
}

function loadSkill(name) {
  var view = document.getElementById("skill-view");
  var tabs = document.querySelectorAll(".skill-tab");
  tabs.forEach(function(t) { t.classList.toggle("active", t.textContent === name); });

  if (skillCache[name]) {
    view.textContent = skillCache[name];
    return;
  }

  view.textContent = "Loading...";
  fetch(RAW_BASE + name + ".md")
    .then(function(r) { return r.ok ? r.text() : Promise.reject("404"); })
    .then(function(text) {
      skillCache[name] = text;
      // only apply if still selected
      var active = document.querySelector(".skill-tab.active");
      if (active && active.textContent === name) {
        view.textContent = text;
      }
    })
    .catch(function() {
      view.textContent = "Failed to load skill: " + name;
    });
}

// ===== Scroll Reveal =====
function initReveal() {
  if (!("IntersectionObserver" in window)) {
    document.querySelectorAll(".reveal").forEach(function(el) { el.classList.add("visible"); });
    return;
  }
  var obs = new IntersectionObserver(function(entries) {
    entries.forEach(function(e) {
      if (e.isIntersecting) { e.target.classList.add("visible"); obs.unobserve(e.target); }
    });
  }, { threshold: 0.08 });
  document.querySelectorAll(".reveal").forEach(function(el) { obs.observe(el); });
}

// ===== Nav scroll =====
function initNav() {
  var nav = document.getElementById("nav");
  if (!nav) return;
  window.addEventListener("scroll", function() {
    nav.classList.toggle("scrolled", window.scrollY > 10);
  }, { passive: true });
}

// ===== TOC =====
function initTOC() {
  var defs = [
    { id: "philosophy", label: "Philosophy" },
    { id: "pipeline", label: "Pipeline" },
    { id: "profiles", label: "Profiles" },
    { id: "enforcement", label: "Enforcement" },
    { id: "skills", label: "Skills" },
    { id: "integration", label: "Integration" },
    { id: "comparison", label: "Comparison" },
    { id: "start", label: "Start" }
  ];
  var toc = document.createElement("div");
  toc.className = "toc";
  defs.forEach(function(d) {
    var a = document.createElement("a");
    a.href = "#" + d.id; a.textContent = d.label; a.dataset.sec = d.id;
    toc.appendChild(a);
  });
  document.body.appendChild(toc);

  function checkWidth() { toc.style.display = window.innerWidth >= 1100 ? "block" : "none"; }
  checkWidth();
  window.addEventListener("resize", checkWidth);

  if (!("IntersectionObserver" in window)) return;
  var obs = new IntersectionObserver(function(entries) {
    entries.forEach(function(e) {
      var link = toc.querySelector('[data-sec="' + e.target.id + '"]');
      if (link) link.classList.toggle("active", e.isIntersecting);
    });
  }, { rootMargin: "-15% 0px -65% 0px" });
  defs.forEach(function(d) {
    var el = document.getElementById(d.id);
    if (el) obs.observe(el);
  });
}

// ===== Init =====
document.addEventListener("DOMContentLoaded", function() {
  initReveal();
  initNav();
  initTOC();
  initSkills();
});
