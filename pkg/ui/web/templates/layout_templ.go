// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.857
package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import (
	"github.com/a-h/templ"
	templruntime "github.com/a-h/templ/runtime"
)

func Base(title string, refreshInterval int) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<!doctype html><html lang=\"en\" class=\"h-full\" x-data=\"{ darkMode: localStorage.getItem(&#39;darkMode&#39;) === &#39;true&#39; }\" :class=\"{ &#39;dark&#39;: darkMode }\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><title>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(title)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `pkg/ui/web/templates/layout.templ`, Line: 9, Col: 17}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, " - SeqCtl</title><script src=\"https://cdn.tailwindcss.com\"></script><script>\n\t\t\t\ttailwind.config = {\n\t\t\t\t\tdarkMode: 'class',\n\t\t\t\t\ttheme: {\n\t\t\t\t\t\textend: {}\n\t\t\t\t\t}\n\t\t\t\t}\n\t\t\t</script><script defer src=\"https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js\"></script><script src=\"https://unpkg.com/htmx.org@1.9.11\"></script><script src=\"https://unpkg.com/htmx.org/dist/ext/ws.js\"></script><script>\n\t\t\t\twindow.SEQCTL_CONFIG = {\n\t\t\t\t\trefreshInterval: { refreshInterval }\n\t\t\t\t};\n\t\t\t</script><style>\n\t\t\t\t.htmx-indicator {\n\t\t\t\t\topacity: 0;\n\t\t\t\t\ttransition: opacity 200ms ease-in;\n\t\t\t\t}\n\t\t\t\t.htmx-request .htmx-indicator {\n\t\t\t\t\topacity: 1;\n\t\t\t\t}\n\t\t\t\t.htmx-request.htmx-indicator {\n\t\t\t\t\topacity: 1;\n\t\t\t\t}\n\t\t\t</style></head><body class=\"h-full bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 transition-colors duration-200\"><div class=\"min-h-full\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = header().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "<main><div class=\"mx-auto max-w-7xl py-6 sm:px-6 lg:px-8\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "</div></main>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = toastContainer().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "</div></body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

func header() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var3 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var3 == nil {
			templ_7745c5c3_Var3 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 6, "<nav class=\"bg-gray-800 dark:bg-gray-950 border-b border-gray-700 dark:border-gray-800\"><div class=\"mx-auto max-w-7xl px-4 sm:px-6 lg:px-8\"><div class=\"flex h-16 items-center justify-between\"><div class=\"flex items-center\"><div class=\"flex-shrink-0\"><a href=\"/\" class=\"text-white font-bold text-xl hover:text-gray-300 transition-colors\">SeqCtl</a></div><div class=\"ml-10 flex items-baseline space-x-4\"><a href=\"/swagger\" class=\"text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium transition-colors\">API Docs</a></div></div><div class=\"flex items-center space-x-4\" x-data=\"{ autoRefresh: true }\"><button @click=\"darkMode = !darkMode; localStorage.setItem(&#39;darkMode&#39;, darkMode)\" class=\"text-gray-300 hover:text-white p-2 rounded-md transition-colors\" title=\"Toggle dark mode\"><svg x-show=\"!darkMode\" class=\"h-5 w-5\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z\"></path></svg> <svg x-show=\"darkMode\" class=\"h-5 w-5\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z\"></path></svg></button> <button @click=\"autoRefresh = !autoRefresh; $dispatch(&#39;toggle-refresh&#39;, { enabled: autoRefresh })\" class=\"p-2 rounded-md transition-colors flex items-center space-x-1\" :class=\"autoRefresh ? &#39;text-green-400 hover:text-green-300&#39; : &#39;text-red-400/70 hover:text-red-400&#39;\" :title=\"autoRefresh ? `Auto-refresh is ON (${window.SEQCTL_CONFIG?.refreshInterval || 5}s)` : &#39;Auto-refresh is OFF&#39;\"><svg x-show=\"autoRefresh\" class=\"h-5 w-5\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15\"></path></svg> <svg x-show=\"!autoRefresh\" class=\"h-5 w-5\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z\"></path></svg> <span x-show=\"autoRefresh\" class=\"text-xs font-medium\" x-text=\"`${window.SEQCTL_CONFIG?.refreshInterval || 5}s`\"></span></button><div class=\"htmx-indicator\"><svg class=\"animate-spin h-5 w-5 text-white\" xmlns=\"http://www.w3.org/2000/svg\" fill=\"none\" viewBox=\"0 0 24 24\"><circle class=\"opacity-25\" cx=\"12\" cy=\"12\" r=\"10\" stroke=\"currentColor\" stroke-width=\"4\"></circle> <path class=\"opacity-75\" fill=\"currentColor\" d=\"M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z\"></path></svg></div></div></div></div></nav>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
