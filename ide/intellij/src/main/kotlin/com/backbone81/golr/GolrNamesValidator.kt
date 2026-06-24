package com.backbone81.golr

import com.intellij.lang.refactoring.NamesValidator
import com.intellij.openapi.project.Project

// Validates names entered in the rename dialog (Shift+F6).
//
// When the user types a new name, IntelliJ calls isIdentifier() on every keystroke.
// If it returns false, the dialog shows a red "not a valid identifier" error and the
// "Refactor" button is disabled, preventing the user from renaming to an illegal name.
//
// isKeyword() guards against renaming to a reserved word.  GoLR has no reserved
// identifiers (keywords are written with a leading "@" and are not valid IDENTIFIER
// tokens), so this always returns false.
//
// This class is registered in plugin.xml as a lang.namesValidator extension.
class GolrNamesValidator : NamesValidator {

    // A GoLR identifier must start with a letter or underscore, followed by any number
    // of letters, digits, or underscores — matching the rule in GolrLexer.readIdentifier().
    override fun isIdentifier(name: String, project: Project?): Boolean =
        name.isNotEmpty()
            && (name[0].isLetter() || name[0] == '_')
            && name.all { it.isLetterOrDigit() || it == '_' }

    // GoLR has no reserved identifier keywords (all keywords start with "@"), so no
    // name can conflict with a keyword.
    override fun isKeyword(name: String, project: Project?): Boolean = false
}
