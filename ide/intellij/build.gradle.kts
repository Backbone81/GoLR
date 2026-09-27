import org.jetbrains.changelog.Changelog
import org.jetbrains.changelog.markdownToHTML
import org.jetbrains.intellij.platform.gradle.TestFrameworkType

plugins {
    id("org.jetbrains.kotlin.jvm")
    id("org.jetbrains.changelog")
    id("org.jetbrains.intellij.platform")
}

// The IntelliJ Platform is compiled for Java 21, so compilation must not run on an older JDK.
kotlin {
    jvmToolchain(21)
}

// Read more: https://plugins.jetbrains.com/docs/intellij/tools-intellij-platform-gradle-plugin.html
dependencies {
    testImplementation(libs.junit)

    // IntelliJ Platform Gradle Plugin Dependencies Extension - read more: https://plugins.jetbrains.com/docs/intellij/tools-intellij-platform-gradle-plugin-dependencies-extension.html
    intellijPlatform {
        // We build, compile and run the tests against the version floor from gradle.properties.
        // Compatibility with newer IDEs is checked statically by verifyPlugin.
        intellijIdea(providers.gradleProperty("platformVersion"))
        testFramework(TestFrameworkType.Platform)

        // Add plugin dependencies for compilation here, for example:
        // bundledPlugin("com.intellij.java")
    }
}

// CHANGELOG.md is written by the changelog utility. The Gradle changelog plugin only reads it for the change notes.
changelog {
    groups.empty()
    // Release headings carry the version with a leading "v", like "## v0.1.0 (2026-09-26)".
    headerParserRegex.set(Regex("""v?(\d+\.\d+\.\d+)"""))
}

intellijPlatform {
    pluginConfiguration {
        // The Marketplace description is README.md without its title.
        description = providers.fileContents(layout.projectDirectory.file("README.md")).asText.map {
            val lines = it.lines()
            val start = lines.indexOfFirst { line -> line.startsWith("# ") }
            val text = if (start >= 0) lines.drop(start + 1).joinToString("\n").trim() else ""
            if (text.isEmpty()) {
                throw GradleException("README.md has no text after the title")
            }
            markdownToHTML(text)
        }

        // The change notes are the changelog section of this version, or the unreleased section before its release.
        // The changelog is read into a local variable, so the lambda does not capture the project, which the
        // configuration cache cannot store.
        val changelog = project.changelog
        changeNotes = providers.gradleProperty("version").map { version ->
            with(changelog) {
                renderItem(
                    (getOrNull(version) ?: getUnreleased()).withHeader(false).withEmptySections(false),
                    Changelog.OutputType.HTML,
                )
            }
        }

        ideaVersion {
            // since-build stays derived from the compile-time platform floor. No until-build: the
            // plugin uses only core platform API (it depends on com.intellij.modules.platform), and
            // verifyPlugin is what actually gates each release against concrete IDE versions.
            untilBuild = provider { null }
        }
    }

    // Static bytecode compatibility check against every currently recommended IDE - the latest
    // release of each supported branch from since-build onward.
    pluginVerification {
        ides {
            recommended()
        }
    }
}
