package openapidoc

import "fmt"

// DocsNavHTML 生成可在 Scalar/Redoc 页面间切换的公共导航。
func DocsNavHTML(active string, scalarPath string, redocPath string) string {
	scalarAttrs := `href="` + scalarPath + `"`
	redocAttrs := `href="` + redocPath + `"`

	if active == "scalar" {
		scalarAttrs += ` aria-current="page"`
	}
	if active == "redoc" {
		redocAttrs += ` aria-current="page"`
	}

	return `<header class="docs-nav">
  <div class="docs-nav__title">Decision Agent API Reference</div>
  <nav class="docs-nav__links">
    <a ` + scalarAttrs + `>Scalar</a>
    <a ` + redocAttrs + `>Redoc</a>
  </nav>
</header>`
}

// DocsPageStyle 返回运行时与静态文档共享的基础样式。
func DocsPageStyle() string {
	return `:root {
      --docs-nav-height: 68px;
      color-scheme: light;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }

    * {
      box-sizing: border-box;
    }

    html {
      background: #f5f7fb;
    }

    body {
      margin: 0;
      padding: var(--docs-nav-height) 0 0;
      background: #f5f7fb;
    }

    .docs-nav {
      position: fixed;
      inset: 0 0 auto 0;
      z-index: 1000;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
      min-height: var(--docs-nav-height);
      padding: 16px 24px;
      background: rgba(15, 23, 42, 0.94);
      color: #f8fafc;
      backdrop-filter: blur(12px);
    }

    .docs-nav__title {
      font-size: 15px;
      font-weight: 600;
      letter-spacing: 0.01em;
    }

    .docs-nav__links {
      display: flex;
      gap: 12px;
      flex-wrap: wrap;
    }

    .docs-nav__links a {
      color: #cbd5e1;
      text-decoration: none;
      padding: 8px 12px;
      border-radius: 999px;
      border: 1px solid rgba(148, 163, 184, 0.32);
      transition: background-color 0.2s ease, color 0.2s ease, border-color 0.2s ease;
    }

    .docs-nav__links a[aria-current="page"] {
      color: #0f172a;
      background: #f8fafc;
      border-color: #f8fafc;
    }

    .scalar-api-references-standalone-search {
      position: sticky;
      top: 0;
      z-index: 4;
      padding-top: 12px;
      background: #f5f7fb;
    }

    .docs-sidebar-group-title {
      margin-top: 18px;
    }

    #redoc-container,
    .scalar-app {
      min-height: calc(100vh - var(--docs-nav-height));
    }

    @media (max-width: 720px) {
      .docs-nav {
        padding: 14px 16px;
        align-items: flex-start;
        flex-direction: column;
      }
    }`
}

// DocsBootstrapScript 在页面中同步导航高度 CSS 变量，避免 Header 遮挡正文或搜索栏。
func DocsBootstrapScript() string {
	return `(() => {
      const root = document.documentElement;

      const syncDocsNavHeight = () => {
        const nav = document.querySelector(".docs-nav");
        if (!nav) {
          return;
        }

        const height = Math.ceil(nav.getBoundingClientRect().height);
        if (height > 0) {
          root.style.setProperty("--docs-nav-height", height + "px");
        }
      };

      syncDocsNavHeight();

      if (typeof ResizeObserver === "function") {
        const nav = document.querySelector(".docs-nav");
        if (!nav) {
          return;
        }

        const observer = new ResizeObserver(syncDocsNavHeight);
        observer.observe(nav);
        window.addEventListener("resize", syncDocsNavHeight);
        window.addEventListener("pagehide", () => observer.disconnect(), { once: true });
        return;
      }

      window.addEventListener("resize", syncDocsNavHeight);
    })();`
}

// ScalarPageEnhancementScript 修正 Scalar 侧栏，把 Models 提升为顶层分组。
func ScalarPageEnhancementScript() string {
	return `(() => {
      const promoteScalarModelsGroup = () => {
        const toc = document.querySelector('nav[aria-label*="Table of contents"] > ul.sidebar-group');
        const modelsItem = toc?.querySelector("#sidebar-models");
        if (!toc || !modelsItem) {
          return;
        }

        let modelsTitle = toc.querySelector('[data-docs-models-group="true"]');
        if (!modelsTitle) {
          modelsTitle = document.createElement("li");
          modelsTitle.className = "sidebar-group-title docs-sidebar-group-title";
          modelsTitle.dataset.docsModelsGroup = "true";
          modelsTitle.textContent = "Models";
        }

        if (modelsTitle.parentElement !== toc) {
          toc.appendChild(modelsTitle);
        }

        if (modelsTitle.nextElementSibling !== modelsItem || modelsItem.parentElement !== toc) {
          toc.appendChild(modelsTitle);
          toc.appendChild(modelsItem);
        }
      };

      const observer = new MutationObserver(promoteScalarModelsGroup);
      observer.observe(document.body, { childList: true, subtree: true });
      promoteScalarModelsGroup();
      window.addEventListener("pagehide", () => observer.disconnect(), { once: true });
    })();`
}

// RedocInitScript 返回带动态滚动偏移的 Redoc 初始化脚本。
func RedocInitScript(specExpr string) string {
	return fmt.Sprintf(`(() => {
      const container = document.getElementById("redoc-container");
      const nav = document.querySelector(".docs-nav");
      const scrollYOffset = Math.ceil(nav?.getBoundingClientRect().height || 68);

      Redoc.init(%s, {
        hideDownloadButton: false,
        scrollYOffset,
      }, container);
    })();`, specExpr)
}
