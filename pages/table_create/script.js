(function () {
  var root = document.getElementById('tableCreateApp');
  if (!root || !window.Vue) return;

  const urlAction = window.urlAction || '/';
  const csrfToken = window.csrfToken || '';

  const { createApp, reactive } = window.Vue;

  createApp({
    setup() {
      const state = reactive({
        schema: '',
        table: '',
        columns: [
          { name: '', type: '', length: '', nullable: false, pk: false, ai: false },
          { name: '', type: '', length: '', nullable: false, pk: false, ai: false },
        ],
        submitting: false,
      });

      function addColumn() {
        state.columns.push({ name: '', type: '', length: '', nullable: false, pk: false, ai: false });
      }
      function removeColumn(idx) {
        if (state.columns.length <= 1) return;
        state.columns.splice(idx, 1);
      }

      async function onSubmit(ev) {
        if (ev && ev.preventDefault) ev.preventDefault();
        try {
          state.submitting = true;
          const form = document.getElementById('tableCreateForm');
          const fd = new FormData(form);
          fd.set('csrf_token', csrfToken);

          const params = new URLSearchParams();
          for (const [k, v] of fd.entries()) params.append(k, v);

          const resp = await fetch(urlAction, {
            method: 'POST',
            body: params,
            credentials: 'same-origin',
            headers: { 'Content-Type': 'application/x-www-form-urlencoded', 'X-CSRF-Token': csrfToken },
          });
          const text = await resp.text();
          let data = null; try { data = text ? JSON.parse(text) : null; } catch (_) {}
          if (resp.ok && data && (data.status === 'success' || data.ok)) {
            window.location.href = window.urlRedirect || urlAction;
            return;
          }
          const msg = (data && (data.message || data.error || data.details)) || text || (`HTTP ${resp.status}`);
          if (window.Swal && window.Swal.fire) window.Swal.fire({ icon: 'error', title: 'Create table failed', text: String(msg).slice(0, 2000) });
          else alert('Create table failed: ' + msg);
        } catch (err) {
          const msg = err && err.message ? err.message : String(err);
          if (window.Swal && window.Swal.fire) window.Swal.fire({ icon: 'error', title: 'Network error', text: String(msg).slice(0, 2000) });
          else alert('Network error: ' + msg);
        } finally {
          state.submitting = false;
        }
      }

      return { ...state, addColumn, removeColumn, onSubmit };
    },
  }).mount('#tableCreateApp');
})();
