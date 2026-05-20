const SITE_CONFIG = {
  doubao: {
    name: '豆包',
    url: 'https://www.doubao.com',
    domains: ['doubao.com']
  },
  deepseek: {
    name: 'DeepSeek',
    url: 'https://chat.deepseek.com',
    domains: ['deepseek.com', 'deepseek.cn']
  },
  qianwen: {
    name: '千问',
    url: 'https://tongyi.aliyun.com/qianwen/',
    domains: ['tongyi.aliyun.com', 'qianwen.aliyun.com']
  },
  kimi: {
    name: 'Kimi',
    url: 'https://kimi.moonshot.cn',
    domains: ['kimi.moonshot.cn', 'moonshot.cn']
  }
};

chrome.runtime.onInstalled.addListener(() => {
  console.log('AI多助手同步工具已安装');
});

chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  switch (message.type) {
    case 'SYNC_CONTENT':
      handleSyncContent(message, sender, sendResponse);
      return true;
    case 'SEND_CONTENT':
      handleSendContent(message, sender, sendResponse);
      return true;
    case 'GET_TABS':
      handleGetTabs(sendResponse);
      return true;
    case 'OPEN_FULLSCREEN':
      handleOpenFullscreen(sendResponse);
      return true;
    case 'REFRESH_TAB':
      handleRefreshTab(message, sendResponse);
      return true;
    case 'CONTENT_SCRIPT_READY':
      console.log('Content script ready from:', sender.tab?.url);
      sendResponse({ success: true });
      return true;
    default:
      return false;
  }
});

async function handleSyncContent(message, sender, sendResponse) {
  try {
    const { content, site } = message;
    const tabs = await getRelevantTabs(site);
    
    const results = [];
    for (const tab of tabs) {
      try {
        const response = await chrome.tabs.sendMessage(tab.id, {
          type: 'SET_INPUT_CONTENT',
          content: content
        });
        results.push({ tabId: tab.id, success: true, response });
      } catch (e) {
        console.warn(`发送到tab ${tab.id} 失败:`, e.message);
        results.push({ tabId: tab.id, success: false, error: e.message });
      }
    }
    
    sendResponse({ success: true, results });
  } catch (error) {
    console.error('同步内容失败:', error);
    sendResponse({ success: false, error: error.message });
  }
}

async function handleSendContent(message, sender, sendResponse) {
  try {
    const { content, site } = message;
    const tabs = await getRelevantTabs(site);
    
    const results = [];
    for (const tab of tabs) {
      try {
        const response = await chrome.tabs.sendMessage(tab.id, {
          type: 'SEND_MESSAGE',
          content: content
        });
        results.push({ tabId: tab.id, success: true, response });
      } catch (e) {
        console.warn(`发送到tab ${tab.id} 失败:`, e.message);
        results.push({ tabId: tab.id, success: false, error: e.message });
      }
    }
    
    sendResponse({ success: true, results });
  } catch (error) {
    console.error('发送内容失败:', error);
    sendResponse({ success: false, error: error.message });
  }
}

async function handleGetTabs(sendResponse) {
  try {
    const allTabs = await chrome.tabs.query({});
    const siteTabs = {};
    
    for (const [siteKey, config] of Object.entries(SITE_CONFIG)) {
      siteTabs[siteKey] = allTabs.filter(tab => {
        return config.domains.some(domain => 
          tab.url && tab.url.includes(domain)
        );
      });
    }
    
    sendResponse({ success: true, tabs: siteTabs });
  } catch (error) {
    console.error('获取标签页失败:', error);
    sendResponse({ success: false, error: error.message });
  }
}

async function handleOpenFullscreen(sendResponse) {
  try {
    const tab = await chrome.tabs.create({
      url: chrome.runtime.getURL('popup.html')
    });
    sendResponse({ success: true, tabId: tab.id });
  } catch (error) {
    console.error('打开全屏模式失败:', error);
    sendResponse({ success: false, error: error.message });
  }
}

async function handleRefreshTab(message, sendResponse) {
  try {
    const { site } = message;
    const tabs = await getRelevantTabs(site);
    
    for (const tab of tabs) {
      await chrome.tabs.reload(tab.id);
    }
    
    sendResponse({ success: true, refreshedCount: tabs.length });
  } catch (error) {
    console.error('刷新标签页失败:', error);
    sendResponse({ success: false, error: error.message });
  }
}

async function getRelevantTabs(site) {
  const allTabs = await chrome.tabs.query({});
  
  if (site && SITE_CONFIG[site]) {
    const config = SITE_CONFIG[site];
    return allTabs.filter(tab => {
      return config.domains.some(domain => 
        tab.url && tab.url.includes(domain)
      );
    });
  }
  
  const allDomains = Object.values(SITE_CONFIG).flatMap(c => c.domains);
  return allTabs.filter(tab => {
    return allDomains.some(domain => 
      tab.url && tab.url.includes(domain)
    );
  });
}

chrome.webNavigation.onCompleted.addListener((details) => {
  if (details.frameId === 0) {
    console.log('页面加载完成:', details.url);
  }
}, {
  url: Object.values(SITE_CONFIG).flatMap(config => 
    config.domains.map(domain => ({ hostContains: domain })
  )
});
