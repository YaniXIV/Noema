(function() {
  var STORAGE_KEY = 'noema_recent_runs';
  var listEl = document.getElementById('recent-runs-list');
  var emptyEl = document.getElementById('recent-runs-empty');

  function loadRecentRuns() {
    try {
      var raw = localStorage.getItem(STORAGE_KEY);
      var runs = raw ? JSON.parse(raw) : [];
      if (runs.length === 0) {
        listEl.innerHTML = '';
        listEl.appendChild(emptyEl);
        emptyEl.style.display = 'block';
        return;
      }
      emptyEl.style.display = 'none';
      listEl.innerHTML = '';
      runs.slice(0, 10).forEach(function(r) {
        var a = document.createElement('a');
        a.href = '/app/results/' + r.run_id;
        a.className = 'recent-run-item';
        a.textContent = (r.name || r.run_id) + ' — ' + r.status + ' · ' + (r.ts ? new Date(r.ts).toLocaleString() : '');
        listEl.appendChild(a);
      });
    } catch (e) {
      listEl.innerHTML = '';
      listEl.appendChild(emptyEl);
      emptyEl.style.display = 'block';
    }
  }

  document.querySelectorAll('.demo-btn').forEach(function(btn) {
    btn.addEventListener('click', function() {
      var demo = this.getAttribute('data-demo');
      alert('Demo "' + demo + '" is client-only. Use New Evaluation for a full run.');
    });
  });

  loadRecentRuns();
})();
