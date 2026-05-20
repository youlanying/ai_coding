(function() {
  'use strict';

  const SITE_CONFIG = {
    doubao: { name: '豆包', url: 'https://www.doubao.com' },
    deepseek: { name: 'DeepSeek', url: 'https://chat.deepseek.com' },
    qianwen: { name: '千问', url: 'https://tongyi.aliyun.com/qianwen/' },
    kimi: { name: 'Kimi', url: 'https://kimi.moonshot.cn' }
  };

  let debounceTimer = null;
  let isSyncEnabled = true;

  const editor = document.getElementById('richEditor');
  const charCount = document.getElementById('charCount');
  const syncStatus = document.getElementById('syncStatus');
  const autoSyncCheckbox = document.getElementById('autoSync');
  const clearBtn = document.getElementById('clearBtn');
  const syncBtn = document.getElementById('syncBtn');
  const sendBtn = document.getElementById('sendBtn');
  const openNewTabBtn = document.getElementById('openNewTab');
  const refreshAllBtn = document.getElementById('refreshAll');
  const iframeRefreshBtns = document.querySelectorAll('.iframe-refresh');

  function init() {
    setupEventListeners();
    restoreEditorContent();
    checkTabs();
  }

  function setupEventListeners() {
    editor.addEventListener('input', handleEditorInput);
    editor.addEventListener('keydown', handleEditorKeydown);
    editor.addEventListener('paste', handlePaste);

    autoSyncCheckbox.addEventListener('change', (e) => {
      isSyncEnabled = e.target.checked;
      updateSyncStatus(isSyncEnabled ? '实时同步已开启' : '实时同步已关闭');
    });

    clearBtn.addEventListener('click', clearEditor);
    syncBtn.addEventListener('click', () => syncContent(false));
    sendBtn.addEventListener('click', () => syncContent(true));

    openNewTabBtn.addEventListener('click', openFullscreen);
    refreshAllBtn.addEventListener('click', refreshAllIframes);

    iframeRefreshBtns.forEach(btn => {
      btn.addEventListener('click', (e) => {
        const site = e.currentTarget.dataset.site;
        refreshIframe(site);
      });
    });

    document.addEventListener('DOMContentLoaded', () => {
      setupPlaceholder();
    });
    setupPlaceholder();
  }

  function setupPlaceholder() {
    editor.addEventListener('focus', () => {
      if (editor.textContent.trim() === '') {
        editor.classList.add('empty');
      }
    });

    editor.addEventListener('blur', () => {
      if (editor.textContent.trim() === '') {
        editor.classList.add('empty');
      } else {
        editor.classList.remove('empty');
      }
    });

    if (editor.textContent.trim() === '') {
      editor.classList.add('empty');
    }
  }

  function handleEditorInput(e) {
    const text = getPlainText();
    updateCharCount(text);
    saveEditorContent(text);

    if (isSyncEnabled) {
      debounceSync();
    } else {
      updateSyncStatus('已暂停实时同步，点击"手动同步"按钮同步');
    }
  }

  function handleEditorKeydown(e) {
    if (e.ctrlKey && e.key === 'Enter') {
      e.preventDefault();
      syncContent(true);
    }
  }

  function handlePaste(e) {
    e.preventDefault();
    const text = e.clipboardData.getData('text/plain');
    document.execCommand('insertText', false, text);
  }

  function getPlainText() {
    return editor.innerText || editor.textContent || '';
  }

  function updateCharCount(text) {
    const count = text.replace(/\s/g, '').length;
    charCount.textContent = `字数: ${count}`;
  }

  function updateSyncStatus(message, isError = false) {
    syncStatus.textContent = message;
    syncStatus.className = 'sync-status' + (isError ? ' error' : '');
  }

  function debounceSync() {
    clearTimeout(debounceTimer);
    updateSyncStatus('正在同步...');
    debounceTimer = setTimeout(() => {
      syncContent(false);
    }, 300);
  }

  async function syncContent(sendAfterSync = false) {
    const content = getPlainText();
    
    if (!content.trim()) {
      updateSyncStatus('请先输入内容');
      return;
    }

    try {
      updateSyncStatus(sendAfterSync ? '正在发送到所有AI助手...' : '正在同步到所有AI助手...');
      
      const messageType = sendAfterSync ? 'SEND_CONTENT' : 'SYNC_CONTENT';
      
      const response = await chrome.runtime.sendMessage({
        type: messageType,
        content: content
      });

      if (response && response.success) {
        const successCount = response.results.filter(r => r.success).length;
        const totalCount = response.results.length;
        
        if (sendAfterSync) {
          updateSyncStatus(`已发送到 ${successCount}/${totalCount} 个AI助手`);
        } else {
          updateSyncStatus(`已同步到 ${successCount}/${totalCount} 个AI助手`);
        }

        if (successCount < totalCount) {
          const failedSites = response.results
            .filter(r => !r.success)
            .map(r => `tab ${r.tabId}`)
            .join(', ');
          console.warn(`部分同步失败: ${failedSites}`);
        }
      } else {
        updateSyncStatus('同步失败，请重试', true);
      }
    } catch (error) {
      console.error('同步错误:', error);
      updateSyncStatus('同步错误: ' + error.message, true);
    }
  }

  function clearEditor() {
    editor.innerHTML = '';
    editor.textContent = '';
    editor.classList.add('empty');
    updateCharCount('');
    updateSyncStatus('已清空');
    
    chrome.runtime.sendMessage({
      type: 'SYNC_CONTENT',
      content: ''
    }).catch(() => {});
  }

  function saveEditorContent(content) {
    try {
      localStorage.setItem('ai_sync_editor_content', content);
    } catch (e) {
      console.warn('保存内容失败:', e);
    }
  }

  function restoreEditorContent() {
    try {
      const saved = localStorage.getItem('ai_sync_editor_content');
      if (saved) {
        editor.textContent = saved;
        editor.classList.remove('empty');
        updateCharCount(saved);
      }
    } catch (e) {
      console.warn('恢复内容失败:', e);
    }
  }

  async function checkTabs() {
    try {
      const response = await chrome.runtime.sendMessage({ type: 'GET_TABS' });
      if (response && response.success) {
        console.log('已打开的标签页:', response.tabs);
      }
    } catch (e) {
      console.warn('检查标签页失败:', e);
    }
  }

  async function openFullscreen() {
    try {
      await chrome.runtime.sendMessage({ type: 'OPEN_FULLSCREEN' });
      window.close();
    } catch (e) {
      console.error('打开全屏模式失败:', e);
      alert('打开全屏模式失败，请手动在新标签页打开');
    }
  }

  function refreshAllIframes() {
    const iframes = document.querySelectorAll('.ai-iframe');
    iframes.forEach(iframe => {
      const src = iframe.src;
      iframe.src = 'about:blank';
      setTimeout(() => {
        iframe.src = src;
      }, 100);
    });
    updateSyncStatus('已刷新所有页面');
  }

  function refreshIframe(site) {
    const iframe = document.getElementById(`iframe-${site}`);
    if (iframe && SITE_CONFIG[site]) {
      const src = SITE_CONFIG[site].url;
      iframe.src = 'about:blank';
      setTimeout(() => {
        iframe.src = src;
      }, 100);
      updateSyncStatus(`已刷新${SITE_CONFIG[site].name}`);
    }
  }

  init();
})();
