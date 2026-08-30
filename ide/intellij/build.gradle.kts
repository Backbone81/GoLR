import groovy.json.JsonSlurper
import org.jetbrains.intellij.platform.gradle.TestFrameworkType
import java.net.URI

plugins {
    id("org.jetbrains.kotlin.jvm")
    id("org.jetbrains.changelog")
    id("org.jetbrains.intellij.platform")
}

// The IntelliJ Platform is compiled for Java 21, so compilation must not run on an older JDK.
kotlin {
    jvmToolchain(21)
}

// Reads the newest IDEA release from the JetBrains release feed. IIU is the product code of the
// unified IDEA distribution, and latest=true narrows the feed to a single entry.
abstract class LatestIdeaVersion : ValueSource<String, ValueSourceParameters.None> {
    override fun obtain(): String {
        val url = "https://data.services.jetbrains.com/products/releases?code=IIU&type=release&latest=true"
        val releases = JsonSlurper().parseText(URI(url).toURL().readText()) as Map<*, *>
        val latest = (releases["IIU"] as? List<*>)?.firstOrNull() as? Map<*, *>
        return latest?.get("version") as? String
            ?: error("Could not read the latest IntelliJ IDEA version from $url")
    }
}

// The platform to build and test against: the floor from gradle.properties, or the newest IDEA
// release with -PplatformLatest=true. The test framework is version locked to the platform, so the
// whole platform moves per run instead of running one build against several IDEs. The latest is a
// moving target on purpose: it warns us when an upcoming IDEA breaks the plugin, at the price of a
// run which can turn red without a change on our side.
val platformVersion = providers.gradleProperty("platformLatest")
    .map { it.toBoolean() }
    .orElse(false)
    .flatMap { latest ->
        when {
            latest -> providers.of(LatestIdeaVersion::class.java) {}
            else -> providers.gradleProperty("platformVersion")
        }
    }

// Read more: https://plugins.jetbrains.com/docs/intellij/tools-intellij-platform-gradle-plugin.html
dependencies {
    testImplementation(libs.junit)

    // IntelliJ Platform Gradle Plugin Dependencies Extension - read more: https://plugins.jetbrains.com/docs/intellij/tools-intellij-platform-gradle-plugin-dependencies-extension.html
    intellijPlatform {
        intellijIdea(platformVersion)
        testFramework(TestFrameworkType.Platform)

        // Add plugin dependencies for compilation here, for example:
        // bundledPlugin("com.intellij.java")
    }
}
