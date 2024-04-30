# Theming

Themes are a set of graphical settings used for consistency within your UI components or application. This document provides an overview of how to use and customize themes.

To get an overview of what is included in a theme you can have a look at [theme.go](./theme.go).

## Predefined Theme
This package includes two predefined themes: *ThemeDefault* and *ThemeMonoSpace*.
It will auto-load *ThemeDefault* theme unless stated otherwise.

## Using Themes
In order to use a theme within your components or anywhere inside your application you should gain access to the defined Theme. You can simply call the GetTheme method

```golang
// will initialize default theme if no one was set before
var theme := ui.GetTheme()
```

You can also set a theme and get it right away like this
```golang
// will initialize default theme if no one was set before
var theme := ui.SetTheme(myFancyThemeInitializer)
```

Then you can use the theme in your application by calling its various methods, or accessing values defined in its Config struct.
```golang
func main() {
  theme := ui.GetTheme()
  fmt.Println(theme.Info("This is an information message"))
  workSucceed := false
  // do some work
  if workSucceed {
    fmt.Printf("%s Work succeed", theme.SuccessIndicator())
  } else {
    fmt.Printf("%s Work Failed", theme.FailureIndicator())
  }
}
```

## Setting a Theme
To set the theme to use you can call
```golang
ui.SetTheme(ui.ThemeMonoSpace)
```

## Define your own theme
In order to create your own theme you have to create a *ThemeInitializer*, which will return a *ThemeConfig*.
You can take inspiration from [themeDefault](themeDefault.go) definition.

For now you **must be exhaustive** and define all fields in ThemeConfig. If this package gains traction, future version will probably inherits from ThemeDefault for undefined fields but this is not the case for now.


