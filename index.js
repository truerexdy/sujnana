async function main() {
  const res = await fetch('blogs.json');
  console.log(res);
  const items = await res.json();
  items.sort((a, b) => new Date(b.updated_time) - new Date(a.updated_time));
  const grid = document.getElementById('blog-grid');
  
  for (const item of items) {
    const link = document.createElement('a');
    link.className = 'blog-card';
    link.href = `/sujnana/blogs/${item.id}.html`;
    link.innerHTML = `<strong>${item.title}</strong><br><small>${formatDate(item.updated_time)}</small>`;
    grid.appendChild(link);
  }
}

function formatDate(iso) {
  const d = new Date(iso);
  return d.toLocaleDateString(undefined, { 
    year: 'numeric', 
    month: 'short', 
    day: 'numeric' 
  });
}

main();
