document.addEventListener('DOMContentLoaded', function() {
    initializeTemplate();
});

function initializeTemplate() {
    const links = document.querySelectorAll('a[href^="#"]');
    links.forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });

    addReadingProgress();
    
    addCodeCopyButtons();
}

function addReadingProgress() {
    const readingSection = document.getElementById('reading-section');
    if (!readingSection) return;

    const progressBar = document.createElement('div');
    progressBar.id = 'reading-progress';
    progressBar.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        width: 0%;
        height: 2px;
        background: #ffffff;
        z-index: 1000;
        transition: width 0.1s ease;
    `;
    document.body.appendChild(progressBar);

    readingSection.addEventListener('scroll', function() {
        const scrollTop = this.scrollTop;
        const scrollHeight = this.scrollHeight - this.clientHeight;
        const progress = (scrollTop / scrollHeight) * 100;
        progressBar.style.width = Math.min(progress, 100) + '%';
    });
}

function addCodeCopyButtons() {
    const codeBlocks = document.querySelectorAll('pre code');
    codeBlocks.forEach(codeBlock => {
        const pre = codeBlock.parentElement;
        pre.style.position = 'relative';
        
        const copyButton = document.createElement('button');
        copyButton.textContent = 'Copy';
        copyButton.style.cssText = `
            position: absolute;
            top: 0.5rem;
            right: 0.5rem;
            background: #333;
            color: #fff;
            border: none;
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            font-size: 0.75rem;
            cursor: pointer;
            opacity: 0.7;
            transition: opacity 0.2s ease;
        `;
        
        copyButton.addEventListener('click', function() {
            navigator.clipboard.writeText(codeBlock.textContent).then(() => {
                copyButton.textContent = 'Copied!';
                setTimeout(() => {
                    copyButton.textContent = 'Copy';
                }, 2000);
            });
        });
        
        copyButton.addEventListener('mouseenter', function() {
            this.style.opacity = '1';
        });
        
        copyButton.addEventListener('mouseleave', function() {
            this.style.opacity = '0.7';
        });
        
        pre.appendChild(copyButton);
    });
}
