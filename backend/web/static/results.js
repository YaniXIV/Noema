(function() {
  var runId = document.body.getAttribute('data-run-id') || document.getElementById('results-run-id').textContent.replace('Run ID: ', '').trim();
  var key = 'noema_run_' + runId;
  var data;
  try {
    var raw = localStorage.getItem(key);
    data = raw ? JSON.parse(raw) : null;
  } catch (e) {
    data = null;
  }

  function copyText(text, button) {
    if (!text) return;
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(function() {
        if (button) showCopied(button);
      });
    } else {
      var temp = document.createElement('textarea');
      temp.value = text;
      document.body.appendChild(temp);
      temp.select();
      document.execCommand('copy');
      document.body.removeChild(temp);
      if (button) showCopied(button);
    }
  }

  function showCopied(button) {
    var original = button.textContent;
    button.textContent = 'Copied';
    button.disabled = true;
    setTimeout(function() {
      button.textContent = original;
      button.disabled = false;
    }, 1200);
  }

  function renderConstraints(list) {
    var container = document.getElementById('results-constraints-list');
    if (!container) return;
    container.innerHTML = '';

    if (!list || !list.length) {
      container.innerHTML =
        '<div class="empty-state">' +
          '<div class="empty-state-title">No constraint summary</div>' +
          '<p class="empty-state-text">This run did not include per-constraint results.</p>' +
        '</div>';
      return;
    }

    function labelSeverity(val) {
      if (val === 0) return 'Limited';
      if (val === 1) return 'Moderate';
      if (val === 2) return 'Severe';
      return '—';
    }

    list.forEach(function(item) {
      var card = document.createElement('div');
      card.className = 'results-constraint-card';
      var title = item.title || item.id || 'Constraint';
      var severity = item.severity !== undefined ? labelSeverity(item.severity) : '—';
      var allowed = item.allowed_max_severity !== undefined ? labelSeverity(item.allowed_max_severity) : '—';
      var verdict = item.pass === true ? 'PASS' : (item.pass === false ? 'FAIL' : '—');
      var verdictClass = item.pass === true ? 'pass' : (item.pass === false ? 'fail' : 'unknown');
      card.setAttribute('data-verdict', verdictClass);
      card.innerHTML =
        '<div class="results-constraint-header">' +
          '<div class="results-constraint-title">' + title + '</div>' +
          '<span class="results-constraint-verdict results-constraint-verdict-' + verdictClass + '">' + verdict + '</span>' +
        '</div>' +
        '<div class="results-constraint-meta">' +
          '<div class="results-constraint-pill">' +
            '<span class="results-constraint-pill-label">Severity</span>' +
            '<span class="results-constraint-pill-value">' + severity + '</span>' +
          '</div>' +
          '<div class="results-constraint-pill">' +
            '<span class="results-constraint-pill-label">Allowed</span>' +
            '<span class="results-constraint-pill-value">' + allowed + '</span>' +
          '</div>' +
        '</div>';
      container.appendChild(card);
    });
  }

  document.getElementById('results-loading').style.display = 'none';
  if (!data) {
    document.getElementById('results-not-found').style.display = 'block';
    return;
  }

  document.getElementById('results-body').style.display = 'block';
  var statusEl = document.getElementById('results-status');
  var status = data.status;
  if (status === true) status = 'PASS';
  if (status === false) status = 'FAIL';
  if (!status && data.public_output) {
    status = data.public_output.overall_pass ? 'PASS' : 'FAIL';
  }
  status = status || '—';
  var statusClass = 'unknown';
  if (status === 'PASS') statusClass = 'pass';
  if (status === 'FAIL') statusClass = 'fail';
  statusEl.className = 'results-status results-status-' + statusClass;
  statusEl.textContent = status;

  var metaEl = document.getElementById('results-summary-meta');
  var metaText = [];
  function labelSeverity(val) {
    if (val === 0) return 'Limited';
    if (val === 1) return 'Moderate';
    if (val === 2) return 'Severe';
    return '—';
  }
  if (data.client && data.client.dataset_source) {
    var sourceLabel = data.client.dataset_source === 'paste' ? 'Pasted JSON' : 'Uploaded file';
    var name = data.client.dataset_name ? (' · ' + data.client.dataset_name) : '';
    metaText.push('Dataset: ' + sourceLabel + name);
  }
  if (data.public_output) {
    if (data.public_output.max_severity !== undefined) metaText.push('Max severity: ' + labelSeverity(data.public_output.max_severity));
    if (data.public_output.policy_threshold !== undefined) metaText.push('Threshold: ' + labelSeverity(data.public_output.policy_threshold));
    if (data.public_output.commitment) metaText.push('Commitment: ' + data.public_output.commitment);
  }
  if (data.verified !== undefined) metaText.push('Verified: ' + (data.verified ? 'Yes' : 'No'));
  metaEl.textContent = metaText.join(' · ');

  var publicPre = document.getElementById('results-public-output-json');
  var publicSection = document.getElementById('results-public-output');
  if (data.public_output) {
    if (publicPre) publicPre.textContent = JSON.stringify(data.public_output, null, 2);
  } else if (publicSection) {
    if (publicPre) publicPre.style.display = 'none';
    var publicEmpty = document.createElement('div');
    publicEmpty.className = 'empty-state';
    publicEmpty.innerHTML =
      '<div class="empty-state-title">No public output stored</div>' +
      '<p class="empty-state-text">This run did not save a public output payload.</p>';
    publicSection.appendChild(publicEmpty);
  }

  var proofSection = document.getElementById('results-proof');
  var proofPre = document.getElementById('results-proof-json');
  var proofMetaEl = document.getElementById('results-proof-meta');
  if (data.proof) {
    if (proofPre) proofPre.textContent = JSON.stringify(data.proof, null, 2);
    var proofMeta = [];
    if (data.proof.system) proofMeta.push('System: ' + data.proof.system);
    if (data.proof.curve) proofMeta.push('Curve: ' + data.proof.curve);
    if (proofMetaEl) proofMetaEl.textContent = proofMeta.join(' · ');
  } else if (proofSection) {
    if (proofPre) proofPre.style.display = 'none';
    if (proofMetaEl) proofMetaEl.textContent = '';
    var proofEmpty = document.createElement('div');
    proofEmpty.className = 'empty-state';
    proofEmpty.innerHTML =
      '<div class="empty-state-title">No proof stored</div>' +
      '<p class="empty-state-text">This run does not have a proof available for download.</p>';
    proofSection.appendChild(proofEmpty);
  }

  var constraints = data.constraint_results || data.constraints || data.per_constraint || [];
  renderConstraints(constraints);

  var copyPublic = document.getElementById('copy-public-output');
  if (copyPublic) {
    if (!data.public_output) {
      copyPublic.textContent = 'No output';
      copyPublic.disabled = true;
    } else {
      copyPublic.addEventListener('click', function() {
        copyText(JSON.stringify(data.public_output || {}, null, 2), copyPublic);
      });
    }
  }

  var copyRunId = document.getElementById('copy-run-id');
  if (copyRunId) {
    copyRunId.addEventListener('click', function() {
      copyText(runId, copyRunId);
    });
  }

  var copyProof = document.getElementById('copy-proof');
  if (copyProof) {
    if (!data.proof) {
      copyProof.textContent = 'No proof';
      copyProof.disabled = true;
    } else {
      copyProof.addEventListener('click', function() {
        copyText(JSON.stringify(data.proof || {}, null, 2), copyProof);
      });
    }
  }

  var copyInputs = document.getElementById('copy-public-inputs');
  if (copyInputs) {
    var inputs = (data.proof && (data.proof.public_inputs_b64 || data.proof.public_inputs)) || (data.public_output && data.public_output.public_inputs) || '';
    if (!inputs) {
      copyInputs.textContent = 'No inputs';
      copyInputs.disabled = true;
    } else {
      copyInputs.addEventListener('click', function() {
        copyText(typeof inputs === 'string' ? inputs : JSON.stringify(inputs, null, 2), copyInputs);
      });
    }
  }
})();
