(function() {
  var input = document.getElementById('file-input');
  var zone = document.getElementById('drop-zone');
  var zoneText = document.getElementById('zone-text');
  var submitBtn = document.getElementById('submit-btn');
  var details = document.getElementById('upload-details');
  var fileName = document.getElementById('upload-file-name');
  var fileSize = document.getElementById('upload-file-size');
  var clearBtn = document.getElementById('clear-file');

  function formatBytes(bytes) {
    if (!bytes && bytes !== 0) return '';
    if (bytes < 1024) return bytes + ' B';
    var kb = bytes / 1024;
    if (kb < 1024) return kb.toFixed(1) + ' KB';
    var mb = kb / 1024;
    if (mb < 1024) return mb.toFixed(1) + ' MB';
    var gb = mb / 1024;
    return gb.toFixed(1) + ' GB';
  }

  function updateZone() {
    if (input.files && input.files.length) {
      var file = input.files[0];
      zoneText.textContent = 'Drop another file to replace';
      if (fileName) fileName.textContent = file.name;
      if (fileSize) fileSize.textContent = formatBytes(file.size);
      if (details) details.style.display = 'flex';
      submitBtn.disabled = false;
    } else {
      zoneText.textContent = 'Drag and drop a file here, or click to choose';
      if (details) details.style.display = 'none';
      submitBtn.disabled = true;
    }
  }

  input.addEventListener('change', updateZone);
  if (clearBtn) {
    clearBtn.addEventListener('click', function() {
      input.value = '';
      zone.classList.remove('dragover');
      updateZone();
    });
  }

  zone.addEventListener('dragover', function(e) {
    e.preventDefault();
    zone.classList.add('dragover');
  });
  zone.addEventListener('dragleave', function() {
    zone.classList.remove('dragover');
  });
  zone.addEventListener('drop', function(e) {
    e.preventDefault();
    zone.classList.remove('dragover');
    if (e.dataTransfer.files.length) {
      input.files = e.dataTransfer.files;
      updateZone();
    }
  });
})();
