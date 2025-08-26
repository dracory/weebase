(function () {
  // Vanilla JS handler for Create Table form
  const root = document.getElementById('tableCreateApp');
  if (!root) return;
  const form = document.getElementById('tableCreateForm');
  if (!form) return;

  const urlAction = window.urlAction || '/';
  const csrfToken = window.csrfToken || '';

  // Add new column row
  const addBtn = document.getElementById('addCol');
  const cols = document.getElementById('cols');
  if (addBtn && cols) {
    addBtn.addEventListener('click', () => {
      const last = cols.querySelector('.col-row');
      const clone = last.cloneNode(true);
      clone.querySelectorAll('input').forEach(i => {
        if (i.type === 'checkbox') { i.checked = false; }
        else { i.value = ''; }
      });
      // Ensure unique checkbox values increment (1..n) so backend maps correctly
      const current = cols.querySelectorAll('.col-row').length + 1;
      clone.querySelectorAll('input[type="checkbox"]').forEach((cb) => { cb.value = String(current); });
      cols.appendChild(clone);
    });
  }

  form.addEventListener('submit', async (e) => {
    e.preventDefault();

    const formData = new FormData(form);
    // append security only; action is already encoded in urlAction
    formData.set('csrf_token', csrfToken);

    // Convert to URLSearchParams for x-www-form-urlencoded
    const params = new URLSearchParams();
    for (const [k, v] of formData.entries()) {
      params.append(k, v);
    }

    try {
      const resp = await fetch(urlAction, {
        method: 'POST',
        body: params,
        credentials: 'same-origin',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded', 'X-CSRF-Token': csrfToken },
      });
      const text = await resp.text();
      let data = null;
      try { data = text ? JSON.parse(text) : null; } catch (_) {}

      if (resp.ok && data && (data.status === 'success' || data.ok)) {
        // On success return to home
        window.location.href = window.urlRedirect || urlAction; // fallback to base
        return;
      }

      const msg = (data && (data.message || data.error || data.details)) || text || (`HTTP ${resp.status}`);
      if (window.Swal && typeof window.Swal.fire === 'function') {
        window.Swal.fire({ icon: 'error', title: 'Create table failed', text: String(msg).slice(0, 2000) });
      } else {
        alert('Create table failed: ' + msg);
      }
    } catch (err) {
      const msg = err && err.message ? err.message : String(err);
      if (window.Swal && typeof window.Swal.fire === 'function') {
        window.Swal.fire({ icon: 'error', title: 'Network error', text: String(msg).slice(0, 2000) });
      } else {
        alert('Network error: ' + msg);
      }
    }
  });
})();
