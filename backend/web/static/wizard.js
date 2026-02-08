(function() {
  var PRESET_CONSTRAINTS = window.NOEMA_PRESET_CONSTRAINTS || [];
  var currentStep = 1;
  var maxStep = 3;
  var customConstraintCount = 0;
  var maxCustomConstraints = 5;
  var datasetSource = 'file';

  function stepEl(n) { return document.querySelector('.stepper-step[data-step="' + n + '"]'); }
  function paneEl(n) { return document.getElementById('pane-' + n); }

  function titleCase(str) {
    return str
      .replace(/_/g, ' ')
      .replace(/\b\w/g, function(m) { return m.toUpperCase(); });
  }

  function escapeHtml(s) {
    var div = document.createElement('div');
    div.textContent = s;
    return div.innerHTML;
  }

  function showStep(step) {
    currentStep = step;
    document.querySelectorAll('.stepper-step').forEach(function(el) {
      var isActive = parseInt(el.getAttribute('data-step'), 10) === step;
      el.classList.toggle('active', isActive);
      el.setAttribute('aria-selected', isActive);
    });
    document.querySelectorAll('.wizard-pane').forEach(function(el) {
      var n = parseInt(el.id.replace('pane-', ''), 10);
      el.classList.toggle('active', n === step);
      el.hidden = n !== step;
    });
    document.getElementById('wizard-prev').disabled = step === 1;
    var nextLabel = 'Next';
    if (step === maxStep - 1) nextLabel = 'Review';
    if (step === maxStep) nextLabel = 'Run evaluation';
    document.getElementById('wizard-next').textContent = nextLabel;
  }

  function setDatasetSource(source) {
    datasetSource = source;
    document.querySelectorAll('.segmented-btn').forEach(function(btn) {
      var active = btn.getAttribute('data-source') === source;
      btn.classList.toggle('active', active);
      btn.setAttribute('aria-selected', active);
    });
    document.querySelectorAll('.dataset-source-pane').forEach(function(pane) {
      var active = pane.getAttribute('data-source') === source;
      pane.classList.toggle('active', active);
      pane.hidden = !active;
    });
    updateDatasetStatus();
  }

  function updateDatasetStatus() {
    var statusEl = document.getElementById('dataset-status');
    var previewEl = document.getElementById('dataset-preview');
    statusEl.classList.remove('is-valid', 'is-invalid');

    if (datasetSource === 'file') {
      var fileInput = document.getElementById('dataset-file');
      if (fileInput.files && fileInput.files.length > 0) {
        var file = fileInput.files[0];
        statusEl.textContent = 'Selected file: ' + file.name + ' (' + Math.round(file.size / 1024) + ' KB)';
        statusEl.classList.add('is-valid');
      } else {
        statusEl.textContent = 'Select a JSON file to continue.';
      }
      previewEl.hidden = true;
      return;
    }

    var paste = document.getElementById('dataset-paste').value.trim();
    if (!paste) {
      statusEl.textContent = 'Paste JSON to continue.';
      previewEl.hidden = true;
      return;
    }

    var parsed = parseJsonSafe(paste);
    if (!parsed.valid) {
      statusEl.textContent = 'Invalid JSON: ' + parsed.error;
      statusEl.classList.add('is-invalid');
      previewEl.hidden = true;
      return;
    }

    statusEl.textContent = 'JSON looks valid.';
    statusEl.classList.add('is-valid');
    var preview = JSON.stringify(parsed.data, null, 2);
    if (preview.length > 700) preview = preview.slice(0, 700) + '\n…';
    previewEl.textContent = preview;
    previewEl.hidden = false;
  }

  function parseJsonSafe(text) {
    try {
      var data = JSON.parse(text);
      return { valid: true, data: data, error: null };
    } catch (err) {
      return { valid: false, data: null, error: err.message || 'Parse error' };
    }
  }

  function buildConstraintsList() {
    var container = document.getElementById('constraints-list');
    container.innerHTML = '';

    var grouped = {};
    PRESET_CONSTRAINTS.forEach(function(c) {
      var key = c.category || 'other';
      if (!grouped[key]) grouped[key] = [];
      grouped[key].push(c);
    });

    Object.keys(grouped).sort().forEach(function(category) {
      var group = document.createElement('div');
      group.className = 'constraint-group';
      group.innerHTML = '<div class="constraint-group-title">' + escapeHtml(titleCase(category)) + '</div>';

      grouped[category].forEach(function(c) {
        var div = document.createElement('div');
        div.className = 'constraint-item';
        var title = titleCase(c.id);
        div.innerHTML =
          '<div class="constraint-row">' +
            '<label class="switch">' +
              '<input type="checkbox" class="constraint-enabled" data-id="' + escapeHtml(c.id) + '">' +
              '<span class="switch-track"></span>' +
            '</label>' +
            '<div>' +
              '<div class="constraint-title">' + escapeHtml(title) + '</div>' +
              '<div class="constraint-subtitle">' + escapeHtml(c.description) + '</div>' +
            '</div>' +
            '<div class="constraint-controls">' +
              '<span class="constraint-subtitle">Max</span>' +
              '<select class="constraint-severity severity-select" data-id="' + escapeHtml(c.id) + '">' +
                '<option value="0">Limited</option><option value="1">Moderate</option><option value="2">Severe</option>' +
              '</select>' +
            '</div>' +
          '</div>' +
          '<details class="constraint-details">' +
            '<summary>Details</summary>' +
            '<div class="constraint-subtitle">ID: ' + escapeHtml(c.id) + '</div>' +
            '<ul class="severity-levels"></ul>' +
          '</details>';

        var ul = div.querySelector('.severity-levels');
        if (c.severity_levels) {
          ['0', '1', '2'].forEach(function(lv) {
            if (c.severity_levels[lv]) {
              var li = document.createElement('li');
              li.textContent = lv + ': ' + c.severity_levels[lv];
              ul.appendChild(li);
            }
          });
        }

        div.querySelector('.constraint-enabled').addEventListener('change', updateConstraintsCount);
        group.appendChild(div);
      });

      container.appendChild(group);
    });

    updateConstraintsCount();
  }

  function updateConstraintsCount() {
    var count = 0;
    document.querySelectorAll('.constraint-enabled').forEach(function(el) {
      if (el.checked) count++;
    });
    var customEnabled = 0;
    document.querySelectorAll('.custom-enabled').forEach(function(el) {
      if (el.checked) customEnabled++;
    });
    var total = count + customEnabled;
    document.getElementById('constraints-count').textContent = total + ' enabled';
  }

  function addCustomConstraintRow() {
    if (customConstraintCount >= maxCustomConstraints) return;
    customConstraintCount++;
    updateCustomLimitHint();

    var div = document.createElement('div');
    div.className = 'custom-constraint-item';
    div.innerHTML =
      '<div class="custom-constraint-fields">' +
        '<input type="text" class="input custom-title" placeholder="Title" maxlength="80">' +
        '<textarea class="input textarea custom-desc" rows="2" placeholder="Description" maxlength="240"></textarea>' +
        '<div class="custom-constraint-actions">' +
          '<div class="custom-constraint-meta">' +
            '<label class="label">Allowed max severity</label>' +
            '<select class="custom-severity severity-select"><option value="0">Limited</option><option value="1">Moderate</option><option value="2">Severe</option></select>' +
            '<label class="checkbox-label"><input type="checkbox" class="custom-enabled" checked> Enabled</label>' +
          '</div>' +
          '<button type="button" class="btn btn-ghost btn-sm remove-custom">Remove</button>' +
        '</div>' +
        '<div class="custom-constraint-error" style="display:none;">Title and description are required.</div>' +
      '</div>';

    var list = document.getElementById('custom-constraints-list');
    list.appendChild(div);

    div.querySelectorAll('.custom-title, .custom-desc, .custom-enabled').forEach(function(el) {
      el.addEventListener('input', function() { validateCustomConstraint(div); });
      el.addEventListener('change', function() { validateCustomConstraint(div); updateConstraintsCount(); });
    });

    div.querySelector('.remove-custom').addEventListener('click', function() {
      div.remove();
      customConstraintCount--;
      updateCustomLimitHint();
      updateConstraintsCount();
    });

    validateCustomConstraint(div);
    updateConstraintsCount();
  }

  function updateCustomLimitHint() {
    var btn = document.getElementById('add-custom-constraint');
    var hint = document.getElementById('custom-constraints-hint');
    btn.disabled = customConstraintCount >= maxCustomConstraints;
    if (customConstraintCount >= maxCustomConstraints) {
      hint.textContent = 'Limit reached: remove one to add more.';
    } else {
      hint.textContent = 'Add up to five custom constraints.';
    }
  }

  function validateCustomConstraint(item) {
    var title = item.querySelector('.custom-title').value.trim();
    var desc = item.querySelector('.custom-desc').value.trim();
    var enabled = item.querySelector('.custom-enabled').checked;
    var errorEl = item.querySelector('.custom-constraint-error');

    if (!enabled) {
      errorEl.style.display = 'none';
      item.querySelector('.custom-title').classList.remove('is-invalid');
      item.querySelector('.custom-desc').classList.remove('is-invalid');
      return true;
    }

    var valid = title.length > 0 && desc.length > 0;
    errorEl.style.display = valid ? 'none' : 'block';
    item.querySelector('.custom-title').classList.toggle('is-invalid', !valid && title.length === 0);
    item.querySelector('.custom-desc').classList.toggle('is-invalid', !valid && desc.length === 0);
    return valid;
  }

  function buildSpec() {
    var spec = {
      schema_version: 1,
      evaluation_name: document.getElementById('evaluation-name').value.trim() || '',
      policy: {
        reveal: {
          max_severity: document.getElementById('reveal-max-severity').checked,
          commitment: document.getElementById('reveal-commitment').checked
        }
      },
      constraints: [],
      custom_constraints: []
    };

    document.querySelectorAll('.constraint-item').forEach(function(item) {
      var id = item.querySelector('.constraint-enabled').getAttribute('data-id');
      var enabled = item.querySelector('.constraint-enabled').checked;
      var severity = parseInt(item.querySelector('.constraint-severity').value, 10);
      spec.constraints.push({ id: id, enabled: enabled, allowed_max_severity: severity });
    });

    document.querySelectorAll('.custom-constraint-item').forEach(function(item) {
      var title = (item.querySelector('.custom-title') && item.querySelector('.custom-title').value) || '';
      var desc = (item.querySelector('.custom-desc') && item.querySelector('.custom-desc').value) || '';
      var severity = parseInt(item.querySelector('.custom-severity').value, 10);
      var enabled = item.querySelector('.custom-enabled').checked;
      spec.custom_constraints.push({
        id: 'custom_' + Date.now() + '_' + Math.random().toString(36).slice(2, 8),
        title: title.trim(),
        description: desc.trim(),
        enabled: enabled,
        allowed_max_severity: severity
      });
    });

    return spec;
  }

  function getDatasetFile() {
    if (datasetSource === 'file') {
      var fileInput = document.getElementById('dataset-file');
      if (fileInput.files && fileInput.files.length > 0) return fileInput.files[0];
      return null;
    }

    var paste = document.getElementById('dataset-paste').value.trim();
    if (paste) {
      var parsed = parseJsonSafe(paste);
      if (parsed.valid) {
        return new File([paste], 'dataset.json', { type: 'application/json' });
      }
      return null;
    }
    return null;
  }

  function updateReviewSummary() {
    var name = document.getElementById('evaluation-name').value.trim() || '(unnamed)';
    var datasetFile = getDatasetFile();
    var datasetSourceLabel = datasetSource === 'file' ? 'Uploaded file' : 'Pasted JSON';
    var datasetSource = datasetFile ? (datasetFile.name || datasetSourceLabel) : '—';
    var spec = buildSpec();
    var enabledCount = spec.constraints.filter(function(c) { return c.enabled; }).length +
      spec.custom_constraints.filter(function(c) { return c.enabled; }).length;

    document.getElementById('review-summary').innerHTML =
      '<p><strong>Evaluation name:</strong> ' + escapeHtml(name) + '</p>' +
      '<p><strong>Dataset:</strong> ' + escapeHtml(datasetSource) + '</p>' +
      '<p><strong>Reveal:</strong> max_severity=' + spec.policy.reveal.max_severity + ', commitment=' + spec.policy.reveal.commitment + '</p>' +
      '<p><strong>Constraints enabled:</strong> ' + enabledCount + '</p>';
  }

  document.querySelectorAll('.segmented-btn').forEach(function(btn) {
    btn.addEventListener('click', function() {
      setDatasetSource(this.getAttribute('data-source'));
    });
  });

  document.getElementById('dataset-file').addEventListener('change', updateDatasetStatus);
  document.getElementById('dataset-paste').addEventListener('input', updateDatasetStatus);

  document.getElementById('wizard-next').addEventListener('click', function() {
    if (currentStep < maxStep) {
      if (currentStep === 1) {
        var df = getDatasetFile();
        if (!df) {
          updateDatasetStatus();
          document.getElementById('dataset-status').classList.add('is-invalid');
          return;
        }
      }
      var next = currentStep + 1;
      showStep(next);
      if (next === 3) updateReviewSummary();
      return;
    }
    submitEvaluation();
  });

  document.getElementById('wizard-prev').addEventListener('click', function() {
    if (currentStep > 1) showStep(currentStep - 1);
  });

  function submitEvaluation() {
    var form = document.getElementById('eval-form');
    if (form.requestSubmit) {
      form.requestSubmit();
      return;
    }
    var evt = new Event('submit', { bubbles: true, cancelable: true });
    form.dispatchEvent(evt);
  }

  document.querySelectorAll('.stepper-step').forEach(function(btn) {
    btn.addEventListener('click', function() {
      var step = parseInt(this.getAttribute('data-step'), 10);
      if (step <= currentStep || step === 1) showStep(step);
    });
  });

  document.getElementById('add-custom-constraint').addEventListener('click', function() {
    addCustomConstraintRow();
  });

  document.getElementById('eval-form').addEventListener('submit', function(e) {
    e.preventDefault();
    var datasetFile = getDatasetFile();
    if (!datasetFile) {
      document.getElementById('submit-error').textContent = 'Please provide a dataset (file or pasted JSON).';
      document.getElementById('submit-error').style.display = 'block';
      return;
    }

    var invalidCustom = false;
    document.querySelectorAll('.custom-constraint-item').forEach(function(item) {
      if (!validateCustomConstraint(item)) invalidCustom = true;
    });

    if (invalidCustom) {
      document.getElementById('submit-error').textContent = 'Please complete required custom constraint fields or disable them.';
      document.getElementById('submit-error').style.display = 'block';
      return;
    }

    document.getElementById('submit-error').style.display = 'none';
    document.getElementById('running-state').style.display = 'block';
    var runBtn = document.getElementById('run-eval-btn');
    if (runBtn) runBtn.disabled = true;
    document.getElementById('wizard-next').disabled = true;

    var spec = buildSpec();
    var formData = new FormData();
    formData.append('spec', JSON.stringify(spec));
    formData.append('dataset', datasetFile);

    var imagesInput = document.getElementById('images-file');
    if (imagesInput && imagesInput.files && imagesInput.files.length > 0) {
      for (var i = 0; i < imagesInput.files.length; i++) {
        formData.append('images', imagesInput.files[i]);
      }
    }

    fetch('/api/evaluate', {
      method: 'POST',
      body: formData,
      credentials: 'same-origin'
    })
      .then(function(res) {
        if (!res.ok) return res.json().then(function(j) { throw new Error(j.error || res.statusText); });
        return res.json();
      })
      .then(function(data) {
        var runId = data.run_id;
        var storageKey = 'noema_run_' + runId;
        try {
          data.client = {
            dataset_source: datasetSource,
            dataset_name: datasetFile && datasetFile.name ? datasetFile.name : '',
            dataset_size: datasetFile && datasetFile.size ? datasetFile.size : 0
          };
          localStorage.setItem(storageKey, JSON.stringify(data));
          var recent = JSON.parse(localStorage.getItem('noema_recent_runs') || '[]');
          recent.unshift({
            run_id: runId,
            status: data.status,
            ts: Date.now(),
            name: spec.evaluation_name || runId
          });
          localStorage.setItem('noema_recent_runs', JSON.stringify(recent.slice(0, 50)));
        } catch (err) {}
        window.location.href = '/app/results/' + encodeURIComponent(runId);
      })
      .catch(function(err) {
        document.getElementById('submit-error').textContent = err.message || 'Evaluation failed.';
        document.getElementById('submit-error').style.display = 'block';
        document.getElementById('running-state').style.display = 'none';
        if (runBtn) runBtn.disabled = false;
        document.getElementById('wizard-next').disabled = false;
      });
  });

  buildConstraintsList();
  setDatasetSource('file');
  updateCustomLimitHint();
})();
