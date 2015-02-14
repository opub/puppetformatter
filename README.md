# puppetformatter
A simple Go program to format your Puppet files based on the [Puppet Language Style Guide](https://docs.puppetlabs.com/guides/style_guide.html).

This was originally just going to be a Go port of [Puppet::Tidy](http://puppetlabs.com/presentations/clean-manifests-puppettidy) that was written in Perl.  It quickly evolved into more than that.  While it doesn't attempt to handle all of the quote replacement that Puppet::Tidy does it has additional features like cleaning up indentation and aligning rockets (=>).

## Features

* Formats an individual *.pp file or all such files within a directory tree
* Rewrites files in place
* Removes trailing whitespace
* Converts tabs to double spaces
* Aligns indentation, nesting appropriately within braces
* Ensures empty newlines before resources
* Replaces C/C++ styles comments (// and /*...*/) with hashed comments
* Removes quotes from standalone variables
* Replaces double quotes with single quotes (when straightforward)
* Aligns rockets (=>) within a block removing any extra whitespace

## TODO

* Quote replacement could be better.  There are so many special cases with both double and single quotes that could be handled.
* File selection is only one *.pp or all *.pp.  There is no mechanism for ignoring files or directories, or handling other file extensions.

## License

See [LICENSE](https://github.com/opub/puppetformatter/blob/master/LICENSE).
