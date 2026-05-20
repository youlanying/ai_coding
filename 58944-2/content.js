(function() {
  'use strict';

  const SITE_SELECTORS = {
    doubao: {
      name: '豆包',
      inputSelectors: [
        'textarea[placeholder*="输入"]',
        'textarea[placeholder*="发送"]',
        'textarea[class*="input"]',
        'div[contenteditable="true"][class*="input"]',
        'div[contenteditable="true"][class*="editor"]',
        'textarea',
        'div[contenteditable="true"]'
      ],
      sendButtonSelectors: [
        'button[aria-label*="发送"]',
        'button[type="submit"]',
        'button[class*="send"]',
        'button:has(svg)',
        'button'
      ]
    },
    deepseek: {
      name: 'DeepSeek',
      inputSelectors: [
        'textarea[placeholder*="输入"]',
        'textarea[placeholder*="发送"]',
        'div[contenteditable="true"][class*="input"]',
        'div[contenteditable="true"][class*="editor"]',
        'textarea',
        'div[contenteditable="true"]'
      ],
      sendButtonSelectors: [
        'button[aria-label*="发送"]',
        'button[type="submit"]',
        'button[class*="send"]',
        'svg[class*="send"]',
        'button'
      ]
    },
    qianwen: {
      name: '千问',
      inputSelectors: [
        'textarea[placeholder*="输入"]',
        'textarea[placeholder*="发送"]',
        'div[contenteditable="true"][class*="input"]',
        'div[contenteditable="true"][class*="editor"]',
        'textarea',
        'div[contenteditable="true"]'
      ],
      sendButtonSelectors: [
        'button[aria-label*="发送"]',
        'button[type="submit"]',
        'button[class*="send"]',
        'button'
      ]
    },
    kimi: {
      name: 'Kimi',
      inputSelectors: [
        'textarea[placeholder*="输入"]',
        'textarea[placeholder*="发送"]',
        'div[contenteditable="true"][class*="input"]',
        'div[contenteditable="true"][class*="editor"]',
        'textarea',
        'div[contenteditable="true"]'
      ],
      sendButtonSelectors: [
        'button[aria-label*="发送"]',
        'button[type="submit"]',
        'button[class*="send"]',
        'svg[class*="send"]',
        'button'
      ]
    }
  };

  function detectSite() {
    const host = window.location.hostname;
    if (host.includes('doubao.com')) return 'doubao';
    if (host.includes('deepseek.com') || host.includes('deepseek.cn')) return 'deepseek';
    if (host.includes('tongyi.aliyun.com') || host.includes('qianwen.aliyun.com')) return 'qianwen';
    if (host.includes('kimi.moonshot.cn') || host.includes('moonshot.cn')) return 'kimi';
    return null;
  }

  const siteKey = detectSite();
  if (!siteKey) return;

  const config = SITE_SELECTORS[siteKey];

  function findElement(selectors, root = document) {
    for (const selector of selectors) {
      try {
        const elements = root.querySelectorAll(selector);
        for (const el of elements) {
          if (isVisible(el)) {
            return el;
          }
        }
      } catch (e) {
        continue;
      }
    }
    return null;
  }

  function isVisible(element) {
    if (!element) return false;
    const style = window.getComputedStyle(element);
    if (style.display === 'none' || style.visibility === 'hidden' || style.opacity === '0') {
      return false;
    }
    const rect = element.getBoundingClientRect();
    return rect.width > 0 && rect.height > 0;
  }

  function findInputElement() {
    const inputEl = findElement(config.inputSelectors);
    if (inputEl) return inputEl;

    const iframes = document.querySelectorAll('iframe');
    for (const iframe of iframes) {
      try {
        const iframeDoc = iframe.contentDocument || iframe.contentWindow.document;
        const iframeInput = findElement(config.inputSelectors, iframeDoc);
        if (iframeInput) return iframeInput;
      } catch (e) {
        continue;
      }
    }

    return null;
  }

  function findSendButton(inputEl) {
    const button = findElement(config.sendButtonSelectors);
    if (button) return button;

    if (inputEl && inputEl.parentElement) {
      const parentButton = inputEl.parentElement.querySelector('button');
      if (parentButton) return parentButton;
    }

    return null;
  }

  function setInputContent(element, content) {
    if (!element) return false;

    try {
      if (element.tagName === 'TEXTAREA' || element.tagName === 'INPUT') {
        element.value = content;
        element.dispatchEvent(new Event('input', { bubbles: true }));
        element.dispatchEvent(new Event('change', { bubbles: true }));
        element.dispatchEvent(new Event('keydown', { bubbles: true }));
        element.dispatchEvent(new Event('keyup', { bubbles: true }));
      } else if (element.isContentEditable) {
        element.innerHTML = escapeHtml(content).replace(/\n/g, '<br>');
        element.dispatchEvent(new Event('input', { bubbles: true }));
        element.dispatchEvent(new Event('change', { bubbles: true }));
        
        const selection = window.getSelection();
        const range = document.createRange();
        range.selectNodeContents(element);
        range.collapse(false);
        selection.removeAllRanges();
        selection.addRange(range);
      } else {
        element.setAttribute('value', content);
        element.textContent = content;
      }

      return true;
    } catch (e) {
      console.error(`设置${config.name}输入框内容失败:`, e);
      return false;
    }
  }

  function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  function clickSendButton(button) {
    if (!button) return false;

    try {
      button.click();
      button.dispatchEvent(new Event('click', { bubbles: true }));
      return true;
    } catch (e) {
      console.error(`点击${config.name}发送按钮失败:`, e);
      return false;
    }
  }

  function sendWithKeyboard(inputEl) {
    if (!inputEl) return false;

    try {
      inputEl.focus();
      const enterEvent = new KeyboardEvent('keydown', {
        key: 'Enter',
        code: 'Enter',
        which: 13,
        keyCode: 13,
        bubbles: true,
        ctrlKey: false,
        metaKey: false,
        shiftKey: false
      });
      inputEl.dispatchEvent(enterEvent);
      
      const keyupEvent = new KeyboardEvent('keyup', {
        key: 'Enter',
        code: 'Enter',
        which: 13,
        keyCode: 13,
        bubbles: true
      });
      inputEl.dispatchEvent(keyupEvent);
      
      return true;
    } catch (e) {
      console.error(`发送键盘事件到${config.name}失败:`, e);
      return false;
    }
  }

  chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    switch (message.type) {
      case 'SET_INPUT_CONTENT':
        handleSetContent(message.content, sendResponse);
        return true;
      case 'SEND_MESSAGE':
        handleSendMessage(message.content, sendResponse);
        return true;
      default:
        return false;
    }
  });

  function handleSetContent(content, sendResponse) {
    const inputEl = findInputElement();
    if (!inputEl) {
      sendResponse({ success: false, error: '未找到输入框' });
      return;
    }

    const success = setInputContent(inputEl, content);
    sendResponse({
      success: success,
      site: siteKey,
      siteName: config.name,
      contentLength: content.length
    });
  }

  function handleSendMessage(content, sendResponse) {
    const inputEl = findInputElement();
    if (!inputEl) {
      sendResponse({ success: false, error: '未找到输入框' });
      return;
    }

    const contentSet = setInputContent(inputEl, content);
    if (!contentSet) {
      sendResponse({ success: false, error: '设置内容失败' });
      return;
    }

    setTimeout(() => {
      const sendButton = findSendButton(inputEl);
      let clicked = false;

      if (sendButton) {
        clicked = clickSendButton(sendButton);
      }

      if (!clicked) {
        clicked = sendWithKeyboard(inputEl);
      }

      sendResponse({
        success: clicked,
        site: siteKey,
        siteName: config.name,
        method: sendButton ? 'button' : 'keyboard'
      });
    }, 100);
  }

  try {
    chrome.runtime.sendMessage({ type: 'CONTENT_SCRIPT_READY', site: siteKey });
  } catch (e) {
    console.log('Content script ready');
  }

  console.log(`[AI助手同步] ${config.name} content script 已加载`);
})();
