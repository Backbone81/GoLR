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

intellijPlatform {
    pluginConfiguration {
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
