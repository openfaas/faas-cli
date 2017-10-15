version=$(curl -s https://api.github.com/repos/openfaas/faas-cli/releases/latest | grep 'tag_name' | cut -d\" -f4)

hasCli() {
    has=$(which faas-cli)

    if [ "$?" = "0" ]; then
        echo "You already have the faas-cli!"
        export n=1
        echo "Overwriting in $n seconds.. Press Control+C to cancel."
        sleep $n
    fi

    hasCurl=$(which curl)
    if [ "$?" = "1" ]; then
        echo "You need curl to use this script."
        exit 1
    fi
}

getPackage() {
    uname=$(uname)

    suffix=""
    temp="/tmp/faas-cli"
    case $uname in
    "Darwin")
    suffix="-darwin"
    ;;
    "Linux")
        arch=$(uname -m)
        echo $arch
        case $arch in
        "aarch64")
        suffix="-arm64"
        ;;
        esac
        case $arch in
        "armv6l" | "armv7l")
        suffix="-armhf"
        ;;
        esac
    ;;
    *_NT*)
    suffix=".exe"
    temp=$temp$suffix
    ;;
    esac

    if [ -e $temp ]; then
        rm $temp
    fi

    url=https://github.com/openfaas/faas-cli/releases/download/$version/faas-cli$suffix
    echo "Getting package $url"

    curl -sSL $url > $temp

    if [ "$?" = "0" ]; then
        echo "Attemping to move faas-cli to /usr/local/bin"
        chmod +x $temp
        cp $temp /usr/local/bin/
        if [ "$?" = "0" ]; then
            echo "New version of faas-cli installed to /usr/local/bin"
        fi
    fi
}

hasCli
getPackage
