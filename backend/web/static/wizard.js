(function() {
  var PRESET_CONSTRAINTS = window.NOEMA_PRESET_CONSTRAINTS || [];
  var currentStep = 1;
  var maxStep = 3;
  var customConstraintCount = 0;
  var maxCustomConstraints = 5;
  var MAX_DATASET_BYTES = 50 * 1024 * 1024;
  var MAX_IMAGE_BYTES = 5 * 1024 * 1024;
  var MAX_IMAGE_COUNT = 10;

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

  function formatBytes(bytes) {
    if (!bytes || bytes <= 0) return '0 B';
    var units = ['B', 'KB', 'MB', 'GB'];
    var idx = Math.floor(Math.log(bytes) / Math.log(1024));
    var value = bytes / Math.pow(1024, idx);
    return value.toFixed(value >= 10 || idx === 0 ? 0 : 1) + ' ' + units[idx];
  }

  function isStep1Ready() {
    return !!getDatasetFile() && isImagesValid();
  }

  function isImagesValid() {
    var imagesInput = document.getElementById('images-file');
    var files = imagesInput && imagesInput.files ? Array.prototype.slice.call(imagesInput.files) : [];
    if (files.length === 0) return true;
    if (files.length > MAX_IMAGE_COUNT) return false;
    return !files.some(function(file) { return file.size > MAX_IMAGE_BYTES; });
  }

  function isStep2Ready() {
    var ok = true;
    document.querySelectorAll('.custom-constraint-item').forEach(function(item) {
      if (!validateCustomConstraint(item)) ok = false;
    });
    return ok;
  }

  function updateWizardNavState() {
    var nextBtn = document.getElementById('wizard-next');
    if (!nextBtn) return;
    if (currentStep === 1) {
      nextBtn.disabled = !isStep1Ready();
      return;
    }
    if (currentStep === 2) {
      nextBtn.disabled = !isStep2Ready();
      return;
    }
    nextBtn.disabled = false;
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
    updateWizardNavState();
  }

  function updateDatasetStatus() {
    var statusEl = document.getElementById('dataset-status');
    statusEl.classList.remove('is-valid', 'is-invalid');
    var fileInput = document.getElementById('dataset-file');
    var zoneText = document.getElementById('dataset-zone-text');
    var details = document.getElementById('dataset-details');
    var fileName = document.getElementById('dataset-file-name');
    var fileSize = document.getElementById('dataset-file-size');
    var clientError = document.getElementById('dataset-client-error');
    var maxBytes = parseInt(fileInput.dataset.maxBytes || '', 10);
    var maxBytesLabel = fileInput.dataset.maxBytesLabel || '';
    if (fileInput) fileInput.classList.remove('is-invalid');

    if (fileInput.files && fileInput.files.length > 0) {
      var file = fileInput.files[0];
      zoneText.textContent = 'Drop another file to replace';
      if (fileName) fileName.textContent = file.name;
      if (fileSize) fileSize.textContent = formatBytes(file.size);
      if (details) details.style.display = 'flex';
      if (Number.isFinite(maxBytes) && maxBytes > 0 && file.size > maxBytes) {
        statusEl.textContent = 'File too large. Max size is ' + maxBytesLabel + '.';
        statusEl.classList.add('is-invalid');
        fileInput.classList.add('is-invalid');
        if (clientError) {
          clientError.textContent = 'File exceeds ' + maxBytesLabel + ' limit.';
          clientError.style.display = 'block';
        }
      } else {
        statusEl.textContent = 'Selected file: ' + file.name + ' (' + formatBytes(file.size) + ')';
        statusEl.classList.add('is-valid');
        if (clientError) {
          clientError.textContent = '';
          clientError.style.display = 'none';
        }
      }
    } else {
      statusEl.textContent = 'Select a JSON file to continue.';
      if (zoneText) zoneText.textContent = 'Drag and drop a JSON file here, or click to choose';
      if (details) details.style.display = 'none';
      if (clientError) {
        clientError.textContent = '';
        clientError.style.display = 'none';
      }
    }
    updateWizardNavState();
  }

  function updateImagesStatus() {
    var statusEl = document.getElementById('images-status');
    if (!statusEl) return;
    statusEl.classList.remove('is-valid', 'is-invalid');

    var imagesInput = document.getElementById('images-file');
    if (imagesInput) imagesInput.classList.remove('is-invalid');
    var files = imagesInput && imagesInput.files ? Array.prototype.slice.call(imagesInput.files) : [];
    if (files.length === 0) {
      statusEl.textContent = 'No images selected.';
      updateWizardNavState();
      return true;
    }

    if (files.length > MAX_IMAGE_COUNT) {
      statusEl.textContent = 'Too many images. Max ' + MAX_IMAGE_COUNT + '.';
      statusEl.classList.add('is-invalid');
      if (imagesInput) imagesInput.classList.add('is-invalid');
      updateWizardNavState();
      return false;
    }

    var oversized = files.find(function(file) { return file.size > MAX_IMAGE_BYTES; });
    if (oversized) {
      statusEl.textContent = 'Image "' + oversized.name + '" exceeds 5MB.';
      statusEl.classList.add('is-invalid');
      if (imagesInput) imagesInput.classList.add('is-invalid');
      updateWizardNavState();
      return false;
    }

    var total = files.reduce(function(sum, file) { return sum + file.size; }, 0);
    statusEl.textContent = 'Images selected: ' + files.length + ' (' + formatBytes(total) + ' total)';
    statusEl.classList.add('is-valid');
    updateWizardNavState();
    return true;
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
      el.addEventListener('input', function() {
        validateCustomConstraint(div);
        updateWizardNavState();
      });
      el.addEventListener('change', function() {
        validateCustomConstraint(div);
        updateConstraintsCount();
        updateWizardNavState();
      });
    });

    div.querySelector('.remove-custom').addEventListener('click', function() {
      div.remove();
      customConstraintCount--;
      updateCustomLimitHint();
      updateConstraintsCount();
      updateWizardNavState();
    });

    validateCustomConstraint(div);
    updateConstraintsCount();
    updateWizardNavState();
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
    updateWizardNavState();
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

  function buildPolicyConfig() {
    var config = {
      policy_version: 'noema_policy_v1',
      constraints: []
    };

    document.querySelectorAll('.constraint-item').forEach(function(item) {
      var id = item.querySelector('.constraint-enabled').getAttribute('data-id');
      var enabled = item.querySelector('.constraint-enabled').checked;
      var severity = parseInt(item.querySelector('.constraint-severity').value, 10);
      config.constraints.push({ id: id, enabled: enabled, max_allowed: severity });
    });

    document.querySelectorAll('.custom-constraint-item').forEach(function(item) {
      var severity = parseInt(item.querySelector('.custom-severity').value, 10);
      var enabled = item.querySelector('.custom-enabled').checked;
      config.constraints.push({
        id: 'custom_' + Date.now() + '_' + Math.random().toString(36).slice(2, 8),
        enabled: enabled,
        max_allowed: severity
      });
    });

    return config;
  }

  function getDatasetFile() {
    var fileInput = document.getElementById('dataset-file');
    if (fileInput.files && fileInput.files.length > 0) {
      var file = fileInput.files[0];
      if (file.size > MAX_DATASET_BYTES) return null;
      return file;
    }
    return null;
  }

  function updateReviewSummary() {
    var name = document.getElementById('evaluation-name').value.trim() || '(unnamed)';
    var datasetFile = getDatasetFile();
    var datasetLabel = datasetFile ? (datasetFile.name || 'Uploaded file') : '—';
    var datasetSize = datasetFile && datasetFile.size ? formatBytes(datasetFile.size) : '';
    var datasetLine = datasetLabel + (datasetSize ? (' · ' + datasetSize) : '');
    var policyConfig = buildPolicyConfig();
    var enabledCount = policyConfig.constraints.filter(function(c) { return c.enabled; }).length;
    var imagesInput = document.getElementById('images-file');
    var images = imagesInput && imagesInput.files ? Array.prototype.slice.call(imagesInput.files) : [];
    var imageTotal = images.reduce(function(sum, file) { return sum + file.size; }, 0);
    var imagesLine = images.length > 0 ? (images.length + ' files · ' + formatBytes(imageTotal)) : 'None';

    document.getElementById('review-summary').innerHTML =
      '<div class="review-item"><span class="review-label">Evaluation</span><span class="review-value">' + escapeHtml(name) + '</span></div>' +
      '<div class="review-item"><span class="review-label">Dataset</span><span class="review-value">' + escapeHtml(datasetLine) + '</span></div>' +
      '<div class="review-item"><span class="review-label">Images</span><span class="review-value">' + escapeHtml(imagesLine) + '</span></div>' +
      '<div class="review-item"><span class="review-label">Constraints</span><span class="review-value">' + enabledCount + ' enabled</span></div>';
  }

  document.getElementById('dataset-file').addEventListener('change', updateDatasetStatus);
  var imagesInputEl = document.getElementById('images-file');
  if (imagesInputEl) imagesInputEl.addEventListener('change', updateImagesStatus);
  var datasetZone = document.getElementById('dataset-drop-zone');
  var datasetClear = document.getElementById('dataset-clear');
  if (datasetClear) {
    datasetClear.addEventListener('click', function() {
      var input = document.getElementById('dataset-file');
      input.value = '';
      if (datasetZone) datasetZone.classList.remove('dragover');
      updateDatasetStatus();
    });
  }
  if (datasetZone) {
    datasetZone.addEventListener('dragover', function(e) {
      e.preventDefault();
      datasetZone.classList.add('dragover');
    });
    datasetZone.addEventListener('dragleave', function() {
      datasetZone.classList.remove('dragover');
    });
    datasetZone.addEventListener('drop', function(e) {
      e.preventDefault();
      datasetZone.classList.remove('dragover');
      var input = document.getElementById('dataset-file');
      if (e.dataTransfer.files.length) {
        input.files = e.dataTransfer.files;
        updateDatasetStatus();
      }
    });
  }

  document.getElementById('wizard-next').addEventListener('click', function() {
    if (currentStep < maxStep) {
      if (currentStep === 1) {
        var df = getDatasetFile();
        if (!df) {
          updateDatasetStatus();
          document.getElementById('dataset-status').classList.add('is-invalid');
          updateWizardNavState();
          return;
        }
        if (!updateImagesStatus()) return;
      }
      if (currentStep === 2 && !isStep2Ready()) return;
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
      document.getElementById('submit-error').textContent = 'Please provide a dataset file.';
      document.getElementById('submit-error').style.display = 'block';
      return;
    }
    if (!updateImagesStatus()) {
      document.getElementById('submit-error').textContent = 'Please fix image upload limits (max 10, 5MB each).';
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

    var policyConfig = buildPolicyConfig();
    var formData = new FormData();
    formData.append('policy_config', JSON.stringify(policyConfig));
    formData.append('dataset_id', datasetFile && datasetFile.name ? datasetFile.name : '');
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
            dataset_source: 'file',
            dataset_name: datasetFile && datasetFile.name ? datasetFile.name : '',
            dataset_size: datasetFile && datasetFile.size ? datasetFile.size : 0
          };
          localStorage.setItem(storageKey, JSON.stringify(data));
          var recent = JSON.parse(localStorage.getItem('noema_recent_runs') || '[]');
          recent.unshift({
            run_id: runId,
            status: data.status || (data.overall_pass ? 'PASS' : 'FAIL'),
            ts: Date.now(),
            name: (document.getElementById('evaluation-name').value.trim() || runId)
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
  updateDatasetStatus();
  updateCustomLimitHint();
  updateImagesStatus();
})();
