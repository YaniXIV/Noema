(function() {
  var STORAGE_KEY = 'noema_recent_runs';
  var MAX_RECENT = 20;
  var listEl = document.getElementById('verify-list');
  var emptyEl = document.getElementById('verify-empty');
  var emptyTitleEl = emptyEl ? emptyEl.querySelector('.empty-state-title') : null;
  var emptyTextEl = emptyEl ? emptyEl.querySelector('.empty-state-text') : null;
  var emptyTitleDefault = emptyTitleEl ? emptyTitleEl.textContent : '';
  var emptyTextDefault = emptyTextEl ? emptyTextEl.textContent : '';
  var filterInput = document.getElementById('verify-filter');
  var countEl = document.getElementById('verify-count');
  var progressEl = document.getElementById('verify-progress');
  var verifyAllBtn = document.getElementById('verify-all');
  var clearBtn = document.getElementById('verify-clear');

  function loadRecentRuns() {
    try {
      var raw = localStorage.getItem(STORAGE_KEY);
      var runs = raw ? JSON.parse(raw) : [];
      if (!Array.isArray(runs)) return [];
      return runs.slice().sort(function(a, b) {
        return (b.ts || 0) - (a.ts || 0);
      });
    } catch (e) {
      return [];
    }
  }

  function loadRunData(runId) {
    try {
      var raw = localStorage.getItem('noema_run_' + runId);
      return raw ? JSON.parse(raw) : null;
    } catch (e) {
      return null;
    }
  }

  function formatTs(ts) {
    if (!ts) return '—';
    try {
      return new Date(ts).toLocaleString();
    } catch (e) {
      return '—';
    }
  }

  function setIndicator(row, state, text) {
    var dot = row.querySelector('.verify-indicator');
    var label = row.querySelector('.verify-indicator-text');
    if (!dot || !label) return;
    row.setAttribute('data-verify', state);
    dot.className = 'verify-indicator verify-indicator-' + state;
    label.textContent = text;
  }

  function copyToClipboard(text, onDone) {
    if (!text) {
      onDone(false);
      return;
    }
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(function() {
        onDone(true);
      }).catch(function() {
        onDone(false);
      });
      return;
    }
    try {
      var ta = document.createElement('textarea');
      ta.value = text;
      ta.style.position = 'fixed';
      ta.style.opacity = '0';
      document.body.appendChild(ta);
      ta.select();
      var ok = document.execCommand('copy');
      document.body.removeChild(ta);
      onDone(ok);
    } catch (e) {
      onDone(false);
    }
  }

  function verifyRun(runId, row, silent) {
    var verifyBtn = row.querySelector('.verify-btn');
    var data = loadRunData(runId);
    if (!data || !data.proof || !data.proof.proof_b64 || !data.proof.public_inputs_b64) {
      setIndicator(row, 'fail', 'Missing proof');
      if (verifyBtn) {
        verifyBtn.disabled = true;
        verifyBtn.textContent = 'Missing proof';
      }
      return Promise.resolve(false);
    }

    if (verifyBtn) verifyBtn.disabled = true;
    setIndicator(row, 'pending', 'Verifying…');

    return fetch('/api/verify', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        run_id: runId,
        proof_b64: data.proof.proof_b64,
        public_inputs_b64: data.proof.public_inputs_b64
      })
    })
      .then(function(res) {
        if (!res.ok) return res.json().then(function(j) { throw new Error(j.error || res.statusText); });
        return res.json();
      })
      .then(function(resp) {
        var ok = !!resp.verified;
        setIndicator(row, ok ? 'ok' : 'fail', ok ? 'Verified' : 'Failed');
        if (verifyBtn) {
          verifyBtn.textContent = ok ? 'Verified' : (silent ? 'Verify' : 'Verify again');
          verifyBtn.disabled = ok;
        }
        return ok;
      })
      .catch(function(err) {
        setIndicator(row, 'fail', err.message || 'Verify failed');
        return false;
      })
      .finally(function() {
        if (verifyBtn && verifyBtn.textContent !== 'Verified') verifyBtn.disabled = false;
      });
  }

  function buildRow(run) {
    var row = document.createElement('div');
    row.className = 'verify-row';
    row.setAttribute('data-run-id', run.run_id);

    var data = loadRunData(run.run_id);

    var main = document.createElement('div');
    main.className = 'verify-main';

    var runLine = document.createElement('div');
    runLine.className = 'verify-runline';

    var runLink = document.createElement('a');
    runLink.className = 'verify-run-id';
    runLink.href = '/verify/' + encodeURIComponent(run.run_id);
    runLink.textContent = run.run_id;

    var status = document.createElement('span');
    var statusVal = (run.status || (data && data.status) || '').toUpperCase();
    var statusClass = statusVal === 'PASS' ? 'pass' : (statusVal === 'FAIL' ? 'fail' : 'unknown');
    status.className = 'verify-status verify-status-' + statusClass;
    status.textContent = statusVal || '—';

    var ts = document.createElement('span');
    ts.className = 'verify-ts';
    ts.textContent = formatTs(run.ts);

    runLine.appendChild(runLink);
    if (run.name && run.name !== run.run_id) {
      var runName = document.createElement('span');
      runName.className = 'verify-run-name';
      runName.textContent = run.name;
      runLine.appendChild(runName);
    }
    runLine.appendChild(status);
    runLine.appendChild(ts);

    var meta = document.createElement('div');
    meta.className = 'verify-subline';

    var indicator = document.createElement('span');
    var indicatorText = document.createElement('span');
    var verified = data && data.verified;
    var state = verified === true ? 'ok' : (verified === false ? 'fail' : 'idle');
    var label = verified === true ? 'Verified' : (verified === false ? 'Failed' : 'Not verified');
    row.setAttribute('data-verify', state);
    indicator.className = 'verify-indicator verify-indicator-' + state;
    indicator.title = label;

    indicatorText.className = 'verify-indicator-text';
    indicatorText.textContent = label;

    meta.appendChild(indicator);
    meta.appendChild(indicatorText);

    main.appendChild(runLine);
    main.appendChild(meta);

    var actions = document.createElement('div');
    actions.className = 'verify-actions';

    var verifyBtn = document.createElement('button');
    verifyBtn.type = 'button';
    verifyBtn.className = 'btn btn-primary btn-sm verify-btn';
    verifyBtn.textContent = verified === true ? 'Verified' : 'Verify';
    if (verified === true) verifyBtn.disabled = true;
    verifyBtn.addEventListener('click', function() {
      verifyRun(run.run_id, row, false);
    });

    var copyProof = document.createElement('button');
    copyProof.type = 'button';
    copyProof.className = 'btn btn-ghost btn-sm';
    copyProof.textContent = 'Copy proof';

    var copyInputs = document.createElement('button');
    copyInputs.type = 'button';
    copyInputs.className = 'btn btn-ghost btn-sm';
    copyInputs.textContent = 'Copy public inputs';

    var proofText = data && data.proof && data.proof.proof_b64 ? data.proof.proof_b64 : '';
    var inputsText = data && data.proof && data.proof.public_inputs_b64 ? data.proof.public_inputs_b64 : '';

    if (!proofText) copyProof.disabled = true;
    if (!inputsText) copyInputs.disabled = true;

    copyProof.addEventListener('click', function() {
      copyToClipboard(proofText, function(ok) {
        copyProof.textContent = ok ? 'Copied' : 'Copy proof';
        setTimeout(function() { copyProof.textContent = 'Copy proof'; }, 1200);
      });
    });

    copyInputs.addEventListener('click', function() {
      copyToClipboard(inputsText, function(ok) {
        copyInputs.textContent = ok ? 'Copied' : 'Copy public inputs';
        setTimeout(function() { copyInputs.textContent = 'Copy public inputs'; }, 1200);
      });
    });

    actions.appendChild(verifyBtn);
    actions.appendChild(copyProof);
    actions.appendChild(copyInputs);

    row.appendChild(main);
    row.appendChild(actions);

    return row;
  }

  function render() {
    var runs = loadRecentRuns();
    var q = (filterInput.value || '').trim().toLowerCase();
    var filtered = runs.filter(function(r) {
      if (!q) return true;
      var runId = (r.run_id || '').toLowerCase();
      var runName = (r.name || '').toLowerCase();
      return runId.indexOf(q) !== -1 || runName.indexOf(q) !== -1;
    });

    listEl.innerHTML = '';
    if (filtered.length === 0) {
      if (q && emptyTitleEl && emptyTextEl) {
        emptyTitleEl.textContent = 'No matches';
        emptyTextEl.textContent = 'Try a different search or clear the filter.';
      } else if (emptyTitleEl && emptyTextEl) {
        emptyTitleEl.textContent = emptyTitleDefault;
        emptyTextEl.textContent = emptyTextDefault;
      }
      emptyEl.style.display = 'block';
      listEl.appendChild(emptyEl);
    } else {
      emptyEl.style.display = 'none';
      filtered.slice(0, MAX_RECENT).forEach(function(r) {
        listEl.appendChild(buildRow(r));
      });
    }

    if (filtered.length) {
      countEl.textContent = 'Showing ' + Math.min(filtered.length, MAX_RECENT) + ' of ' + filtered.length + ' runs';
    } else {
      countEl.textContent = '';
    }
    if (verifyAllBtn) verifyAllBtn.disabled = filtered.length === 0;
    if (clearBtn) clearBtn.disabled = runs.length === 0;
  }

  function verifyAllVisible() {
    var rows = Array.prototype.slice.call(listEl.querySelectorAll('.verify-row'));
    if (rows.length === 0) return;

    verifyAllBtn.disabled = true;
    var total = rows.length;
    var done = 0;

    function updateProgress() {
      progressEl.textContent = 'Verifying ' + done + '/' + total + '…';
    }

    updateProgress();

    var chain = Promise.resolve();
    rows.forEach(function(row) {
      var runId = row.getAttribute('data-run-id');
      chain = chain.then(function() {
        return verifyRun(runId, row, true);
      }).then(function() {
        done++;
        updateProgress();
      });
    });

    chain.finally(function() {
      verifyAllBtn.disabled = false;
      progressEl.textContent = 'Verification complete';
      setTimeout(function() {
        progressEl.textContent = '';
      }, 2000);
    });
  }

  function clearHistory() {
    var runs = loadRecentRuns();
    if (runs.length === 0) return;
    var ok = window.confirm('Clear local verification history? This only affects this browser.');
    if (!ok) return;
    try {
      localStorage.removeItem(STORAGE_KEY);
      runs.forEach(function(r) {
        if (r && r.run_id) localStorage.removeItem('noema_run_' + r.run_id);
      });
    } catch (e) {
      // Best-effort cleanup only.
    }
    progressEl.textContent = 'History cleared';
    setTimeout(function() {
      progressEl.textContent = '';
    }, 2000);
    render();
  }

  if (filterInput) {
    filterInput.addEventListener('input', render);
  }

  if (verifyAllBtn) {
    verifyAllBtn.addEventListener('click', verifyAllVisible);
  }

  if (clearBtn) {
    clearBtn.addEventListener('click', clearHistory);
  }

  render();
})();
