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
        finally{ this.loading = false; this.render(); }
      },
      render(){
        var el = document.getElementById('wb-objects');
        if (!el) return;
        el.innerHTML = '';
        if (this.loading){ el.innerHTML = '<li>loading...</li>'; return; }
        if (this.error){ el.innerHTML = '<li class="text-red-600">'+this.error.replace(/</g,'&lt;')+'</li>'; return; }
        var base = window.urlBrowseRows || '';
        if (this.tables.length === 0){ el.innerHTML = '<li>No tables</li>'; return; }
        this.tables.forEach(function(t){
          var li = document.createElement('li');
          var a = document.createElement('a');
          a.textContent = 'select ' + t;
          a.href = base + (base.indexOf('?')>-1?'&':'?') + 'table=' + encodeURIComponent(t);
          a.className = 'hover:underline';
          li.appendChild(a);
          el.appendChild(li);
        });
      }
    }
  }).mount(document.createElement('div'));
})();
