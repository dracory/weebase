(function(){
  if (!window.Vue) return;
  const { createApp } = window.Vue;
  createApp({
    data(){
      return { tables: [], loading: true, error: '' };
    },
    mounted(){ this.load(); },
    methods:{
      async load(){
        try{
          const url = window.urlListTables || '';
          const r = await fetch(url, { credentials: 'same-origin' });
          const d = await r.json();
          this.tables = (d && d.data && Array.isArray(d.data.tables)) ? d.data.tables : [];
        }catch(e){ this.error = e && e.message ? e.message : String(e); }
        finally{ this.loading = false; this.renderSidebar(); this.renderMain(); }
      },
      renderSidebar(){
        var el = document.getElementById('wb-objects');
        if (!el) return;
        el.innerHTML = '';
        if (this.loading){ el.innerHTML = '<li>loading...</li>'; return; }
        if (this.error) { 
          el.innerHTML = '<li class="text-red-600">' + 
            this.error.toString()
              .replace(/&/g, '&amp;')
              .replace(/</g, '&lt;')
              .replace(/>/g, '&gt;') + 
            '</li>'; 
          return; 
        }
        var base = window.urlTable || window.urlBrowseBase || '';
        if (this.tables.length === 0){ el.innerHTML = '<li>No tables</li>'; return; }
        this.tables.forEach(function(t){
          var li = document.createElement('li');
          var a = document.createElement('a');
          a.textContent = 'select ' + t;
          a.href = base + (base.indexOf('?') > -1 ? '&' : '?') + 'table=' + encodeURIComponent(t);
          a.className = 'hover:underline';
          li.appendChild(a);
          el.appendChild(li);
        });
      },
      renderMain(){
        var container = document.querySelector('.wb-main .wb-container');
        if (!container) return;
        // Remove previous if any
        var existing = container.querySelector('#wb-home-grid');
        if (existing) existing.remove();

        var wrap = document.createElement('section');
        wrap.id = 'wb-home-grid';
        wrap.className = 'mt-2';

        var h = document.createElement('h3');
        h.textContent = 'Tables and views';
        h.className = 'text-lg font-semibold mb-2';
        wrap.appendChild(h);

        // Search bar (client filter)
        var searchWrap = document.createElement('div');
        searchWrap.className = 'mb-2';
        var input = document.createElement('input');
        input.type = 'text';
        input.placeholder = 'Search';
        input.className = 'border rounded px-2 py-1';
        searchWrap.appendChild(input);
        wrap.appendChild(searchWrap);

        var table = document.createElement('table');
        table.className = 'min-w-full text-sm border border-slate-200 dark:border-slate-800';
        var thead = document.createElement('thead');
        thead.innerHTML = '<tr class="bg-slate-50 dark:bg-slate-900"><th class="text-left px-2 py-1 border">Table</th><th class="text-left px-2 py-1 border">Rows</th><th class="px-2 py-1 border">Actions</th></tr>';
        var tbody = document.createElement('tbody');

        var base = window.urlTable || window.urlBrowseBase || '';
        var rows = this.tables.slice();
        function renderRows(filter){
          tbody.innerHTML='';
          rows.forEach(function(t){
            if (filter && t.toLowerCase().indexOf(filter.toLowerCase()) === -1) return;
            var tr = document.createElement('tr');
            var tdName = document.createElement('td'); tdName.className='px-2 py-1 border';
            var a = document.createElement('a'); 
            a.className = 'text-blue-700 hover:underline'; 
            a.textContent = t; 
            a.href = base + (base.indexOf('?') > -1 ? '&' : '?') + 'table=' + encodeURIComponent(t);
            tdName.appendChild(a);
            var tdRows = document.createElement('td'); tdRows.className='px-2 py-1 border'; tdRows.textContent='—';
            var tdAct = document.createElement('td'); tdAct.className='px-2 py-1 border text-center';
            var sel = document.createElement('a'); sel.href=a.href; sel.textContent='select'; sel.className='hover:underline';
            tdAct.appendChild(sel);
            tr.appendChild(tdName); tr.appendChild(tdRows); tr.appendChild(tdAct);
            tbody.appendChild(tr);
          });
        }
        renderRows('');
        input.addEventListener('input', function(){ renderRows(input.value || ''); });

        table.appendChild(thead); table.appendChild(tbody);
        wrap.appendChild(table);
        container.prepend(wrap);
      }
    }
  }).mount(document.createElement('div'));
})();
