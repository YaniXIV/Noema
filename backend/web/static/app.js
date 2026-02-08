(function() {
  var STORAGE_KEY = 'noema_recent_runs';
  var listEl = document.getElementById('recent-runs-list');
  var emptyEl = document.getElementById('recent-runs-empty');

  function statusClass(status) {
    if (!status) return 'status-pill-unknown';
    var normalized = String(status).toUpperCase();
    if (normalized === 'PASS') return 'status-pill-pass';
    if (normalized === 'FAIL') return 'status-pill-fail';
    return 'status-pill-unknown';
  }

  function formatTimestamp(ts) {
    if (!ts) return null;
    var date = new Date(ts);
    if (Number.isNaN(date.getTime())) return null;
    var diffMs = Math.abs(Date.now() - date.getTime());
    var minutes = Math.round(diffMs / 60000);
    var label = '';
    if (minutes < 1) label = 'just now';
    else if (minutes < 60) label = minutes + 'm ago';
    else if (minutes < 1440) label = Math.round(minutes / 60) + 'h ago';
    else label = Math.round(minutes / 1440) + 'd ago';
    return {
      label: label,
      title: date.toLocaleString()
    };
  }

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

        var header = document.createElement('div');
        header.className = 'recent-run-header';

        var name = document.createElement('div');
        name.className = 'recent-run-name';
        name.textContent = r.name || r.run_id;

        var status = document.createElement('span');
        status.className = 'status-pill ' + statusClass(r.status);
        status.textContent = r.status || 'UNKNOWN';

        header.appendChild(name);
        header.appendChild(status);

        var meta = document.createElement('div');
        meta.className = 'recent-run-meta';

        var runId = document.createElement('span');
        runId.textContent = 'Run ID: ' + r.run_id;
        meta.appendChild(runId);

        var ts = formatTimestamp(r.ts);
        if (ts) {
          var time = document.createElement('span');
          time.textContent = ts.label;
          time.title = ts.title;
          meta.appendChild(time);
        }

        a.appendChild(header);
        a.appendChild(meta);
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
