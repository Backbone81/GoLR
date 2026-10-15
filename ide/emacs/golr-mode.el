;;; golr-mode.el --- Major mode for GoLR grammar files -*- lexical-binding: t; -*-

;; URL: https://github.com/backbone81/golr
;; Package-Requires: ((emacs "29.1"))

;;; Commentary:

;; Syntax highlighting and comment syntax for GoLR grammar files (.golr).
;; Language features come from the GoLR language server, see README.md.

;;; Code:

(defvar golr-mode-syntax-table
  (let ((table (make-syntax-table)))
    (modify-syntax-entry ?/ ". 124b" table)
    (modify-syntax-entry ?* ". 23" table)
    (modify-syntax-entry ?\n "> b" table)
    (modify-syntax-entry ?_ "w" table)
    (modify-syntax-entry ?@ "_" table)
    table)
  "Syntax table for `golr-mode'.")

(defun golr-mode--propertize-regex ()
  "Mark the delimiters of the regex literal just matched as string fences.
A match inside a comment or string, or starting or ending a comment, is skipped."
  (let ((start (match-beginning 0))
        (end (match-end 0)))
    (if (or (nth 8 (save-excursion (syntax-ppss start)))
            (nth 8 (save-excursion (syntax-ppss (1+ start))))
            (memq (char-after (1+ start)) '(?/ ?*)))
        (goto-char (1+ start))
      (put-text-property start (1+ start) 'syntax-table (string-to-syntax "|"))
      (put-text-property (1- end) end 'syntax-table (string-to-syntax "|")))))

(defun golr-mode--syntax-propertize (start end)
  "Mark the regular expression literals between START and END.
A regular expression literal ends at the end of its line, as in the
GoLR scanner."
  (goto-char start)
  (funcall
   (syntax-propertize-rules
    ("/\\(?:[^\n/\\]\\|\\\\.\\)*/" (0 (ignore (golr-mode--propertize-regex)))))
   (point) end))

(defun golr-mode--syntactic-face (state)
  "Return the face of the string or comment described by the parse STATE."
  (cond ((nth 4 state) 'font-lock-comment-face)
        ((eq (char-after (nth 8 state)) ?/) 'font-lock-regexp-face)
        (t 'font-lock-string-face)))

(defun golr-mode--escape-face ()
  "Return the escape face if the escape just matched is in a string or regex."
  (when (nth 3 (save-excursion (syntax-ppss (match-beginning 0))))
    'font-lock-escape-face))

(defconst golr-mode-font-lock-keywords
  `((,(regexp-opt '("@scanner" "@parser") 'symbols) 0 'font-lock-keyword-face)
    (,(regexp-opt '("@skip" "@fragment" "@empty" "@error" "@start" "@left" "@right" "@none" "@precedence" "@name")
                  'symbols)
     0 'font-lock-builtin-face)
    ("|" 0 'font-lock-operator-face)
    (";" 0 'font-lock-delimiter-face)
    ("[{}]" 0 'font-lock-bracket-face)
    ("\\\\." 0 (golr-mode--escape-face) prepend))
  "Highlighting for `golr-mode'.")

;;;###autoload
(define-derived-mode golr-mode prog-mode "GoLR"
  "Major mode for GoLR grammar files."
  (setq-local comment-start "// ")
  (setq-local comment-end "")
  (setq-local comment-start-skip "\\(?://+\\|/\\*+\\)\\s-*")
  (setq-local syntax-propertize-function #'golr-mode--syntax-propertize)
  (setq-local font-lock-defaults
              '(golr-mode-font-lock-keywords
                nil nil nil nil
                (font-lock-syntactic-face-function . golr-mode--syntactic-face))))

;;;###autoload
(add-to-list 'auto-mode-alist '("\\.golr\\'" . golr-mode))

(provide 'golr-mode)

;;; golr-mode.el ends here
