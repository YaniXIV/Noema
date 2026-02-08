(function() {
  var input = document.getElementById('file-input');
  var zone = document.getElementById('drop-zone');
  var zoneText = document.getElementById('zone-text');
  var submitBtn = document.getElementById('submit-btn');

  function updateZone() {
    if (input.files && input.files.length) {
      zoneText.textContent = input.files[0].name;
      submitBtn.disabled = false;
    } else {
      zoneText.textContent = 'Drag and drop a file here, or click to choose';
      submitBtn.disabled = true;
    }
  }

  input.addEventListener('change', updateZone);

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
