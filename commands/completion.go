package commands

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var shell string

var completionCmd = &cobra.Command{
	Use:   "completion SHELL",
	Short: "Generates shell auto completion",
	Long: `Generates shell auto completion for Bash or ZSH.

Please follow the instructions in the link below to activate the shell auto completion in your environment:
https://docs.openfaas.com/cli/completion/`,
	Example: `  faas-cli completion --shell bash
  faas-cli completion --shell zsh`,
	RunE: runCompletion,
}

func init() {
	completionCmd.Flags().StringVar(&shell, "shell", "", "Outputs shell completion, must be bash or zsh")
	completionCmd.MarkFlagRequired("shell")

	faasCmd.AddCommand(completionCmd)
}

func runCompletion(cmd *cobra.Command, args []string) (err error) {
	if shell == "" {
		return fmt.Errorf("--shell is required and must be bash or zsh")
	}

	switch shell {
	case "bash":
		err = generateBashCompletion()
		if err != nil {
			return err
		}
		return nil

	case "zsh":
		err = generateZshCompletion()
		if err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("%q shell not supported, must be bash or zsh", shell)
	}
}

func generateBashCompletion() error {
	err := faasCmd.GenBashCompletion(os.Stdout)
	if err != nil {
		return err
	}

	return nil
}

func generateZshCompletion() error {
	zshHead := "#compdef faas-cli\n"

	out := os.Stdout

	_, err := out.Write([]byte(zshHead))
	if err != nil {
		return err
	}

	zshInitialization := `
__faas-cli_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}
__faas-cli_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift
		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__faas-cli_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}
__faas-cli_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?
	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}
__faas-cli_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}
__faas-cli_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}
__faas-cli_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}
__faas-cli_filedir() {
	local RET OLD_IFS w qw
	__faas-cli_debug "_filedir $@ cur=$cur"
	if [[ "$1" = \~* ]]; then
		# somehow does not work. Maybe, zsh does not call this at all
		eval echo "$1"
		return 0
	fi
	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"
	IFS="," __faas-cli_debug "RET=${RET[@]} len=${#RET[@]}"
	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__faas-cli_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}
__faas-cli_quote() {
	if [[ $1 == \'* || $1 == \"* ]]; then
		# Leave out first character
		printf %q "${1:1}"
	else
	printf %q "$1"
	fi
}
autoload -U +X bashcompinit && bashcompinit
# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi
__faas-cli_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__faas-cli_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__faas-cli_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__faas-cli_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__faas-cli_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__faas-cli_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/builtin declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__faas-cli_type/g" \
	<<'BASH_COMPLETION_EOF'
`
	_, err = out.Write([]byte(zshInitialization))
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	faasCmd.GenBashCompletion(buf)

	_, err = out.Write(buf.Bytes())
	if err != nil {
		return err
	}

	zshTail := `
BASH_COMPLETION_EOF
}
__faas-cli_bash_source <(__faas-cli_convert_bash_to_zsh)
_complete faas-cli 2>/dev/null
`
	_, err = out.Write([]byte(zshTail))
	if err != nil {
		return err
	}

	return nil
}
