// ===== Scroll Reveal =====
function initReveal() {
  if (!("IntersectionObserver" in window)) {
    document.querySelectorAll(".reveal").forEach(function(el) {
      el.classList.add("visible");
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

// ===== Staggered step reveals =====
function initStepReveal() {
  if (!("IntersectionObserver" in window)) return;
  var steps = document.querySelectorAll(".step, .ef-step, .skill-card");
  var observer = new IntersectionObserver(function(entries) {
    entries.forEach(function(entry) {
      if (entry.isIntersecting) {
        entry.target.classList.add("visible");
        observer.unobserve(entry.target);
      }
    });
  }, { threshold: 0.05 });
  steps.forEach(function(el, i) {
    el.classList.add("reveal");
    el.style.transitionDelay = (i % 4) * 0.1 + "s";
    observer.observe(el);
  });
}

// ===== Nav scroll =====
function initNav() {
  var nav = document.getElementById("nav");
  if (!nav) return;
  window.addEventListener("scroll", function() {
    nav.classList.toggle("scrolled", window.scrollY > 10);
  }, { passive: true });
}

// ===== Floating TOC =====
function initTOC() {
  var sections = [
    { id: "philosophy", label: "Philosophy" },
    { id: "pipeline", label: "Pipeline" },
    { id: "profiles", label: "Profiles" },
    { id: "enforcement", label: "Enforcement" },
    { id: "skills", label: "Skills" },
    { id: "integration", label: "Integration" },
    { id: "comparison", label: "v1 vs v2" },
    { id: "start", label: "Start" }
  ];

  // build toc
  var toc = document.createElement("div");
  toc.className = "toc";
  sections.forEach(function(s) {
    var a = document.createElement("a");
    a.href = "#" + s.id;
    a.textContent = s.label;
    a.dataset.section = s.id;
    toc.appendChild(a);
  });
  document.body.appendChild(toc);

  // show only on wide screens
  function checkWidth() {
    toc.style.display = window.innerWidth >= 1100 ? "block" : "none";
  }
  checkWidth();
  window.addEventListener("resize", checkWidth);

  // highlight active section
  if (!("IntersectionObserver" in window)) return;
  var observer = new IntersectionObserver(function(entries) {
    entries.forEach(function(entry) {
      var link = toc.querySelector('[data-section="' + entry.target.id + '"]');
      if (link) link.classList.toggle("active", entry.isIntersecting);
    });
  }, { rootMargin: "-20% 0px -60% 0px" });
  sections.forEach(function(s) {
    var el = document.getElementById(s.id);
    if (el) observer.observe(el);
  });
}

// ===== Init =====
document.addEventListener("DOMContentLoaded", function() {
  initReveal();
  initStepReveal();
  initNav();
  initTOC();
});
