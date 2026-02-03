<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://i.imgur.com/QEmYDw4.png">
    <img alt="jqfmt logo" src="https://i.imgur.com/oQTnVNb.png" width="300px">
  </picture>
</p>

## Description

I'm frequently passed long shell one-liners that require some visual inspection before running. These days, there's about as much jq in that one-liner as there is bash. I wrote jqfmt to help add line breaks in sensible locations while reading (or writing!) jq.

<p align="center">
<img alt="ray" src="https://i.imgur.com/sgII1ce.png" width="400px">
</p>

At time of initial development, I naturally turned to https://github.com/itchyny/gojq expecting to be able to generate and walk a syntax tree—but gojq didn't seem to provide an AST that could be "walked," and it doesn't export its parsing logic to be used in library form. So, I yanked the relevant code out of https://github.com/itchyny/gojq/blob/main/query.go and started from there.

Side note: Ever tried Googling for "jq formatter"? Reading search results is a nightmare since jq itself _is_, among other things, a formatter.

## Getting started

### Install

With [Homebrew](https://brew.sh) via the [`jqfmt` package](https://formulae.brew.sh/formula/jqfmt):

```shell
brew install jqfmt
```

With Go:

```shell
go install -v github.com/noperator/jqfmt/cmd/jqfmt@latest
```

### Usage

```
𝄢 jqfmt -h
Usage of jqfmt:
  -ar
    	arrays
  -f string
    	file
  -o	one line
  -ob
    	objects
  -op string
    	operators
  -v	verbose
```

Let's take this line of jq…

```
𝄢 echo '{one: .two, three: [.four, .five, [.fivetwo, .fivethree]], six: map(select((.seven | .eight | .nine)))}' |
    jqfmt
{ one: .two, three: [.four, .five, [.fivetwo, .fivethree]], six: map(select((.seven | .eight | .nine))) }
```

…and format objects.

```

𝄢 echo '{one: .two, three: [.four, .five, [.fivetwo, .fivethree]], six: map(select((.seven | .eight | .nine)))}' |
    jqfmt -ob
{
    one: .two,
    three: [.four, .five, [.fivetwo, .fivethree]],
    six: map(select((.seven | .eight | .nine)))
}
```

Nice! Let's also do arrays.

```
𝄢 echo '{one: .two, three: [.four, .five, [.fivetwo, .fivethree]], six: map(select((.seven | .eight | .nine)))}' |
    jqfmt -ob -ar
{
    one: .two,
    three: [
        .four,
        .five,
        [
            .fivetwo,
            .fivethree
        ]
    ],
    six: map(select((.seven | .eight | .nine)))
}
```

It'll read easier if we also break on pipes.

```
𝄢 echo '{one: .two, three: [.four, .five, [.fivetwo, .fivethree]], six: map(select((.seven | .eight | .nine)))}' |
    jqfmt -ob -ar -op pipe
{
    one: .two,
    three: [
        .four,
        .five,
        [
            .fivetwo,
            .fivethree
        ]
    ],
    six: map(select((.seven |
        .eight |
        .nine)))
}
```

<details><summary>Full list of valid operators</summary>
<p>

```
pipe
comma
add
sub
mul
div
mod
eq
ne
gt
lt
ge
le
and
or
alt
assign
modify
updateAdd
updateSub
updateMul
updateDiv
updateMod
updateAlt
```

</p>
</details>

## Back matter

### Acknowledgements

- @zjzeit for helping me believe that formatting jq is a reasonable thing to do.
- @colindean for helping me believe that jqfmt was a reasonable thing to build.
- @tracertea for writing jq lines long enough to warrant a formatter.
- @addyosmani for "First do it, then do it right, then do it better." I wrote this on a plane (as the [best](https://github.com/tomnomnom/hacks/tree/master/kxss) [tools](https://github.com/moloch--/godns) are) over a year ago, but always had more features I wanted to add before releasing it. Doing it "right" might mean a refactor or clean-up; "better" would probably mean incorporating an AST.

### See also

- https://news.ycombinator.com/item?id=9448128
- https://github.com/stedolan/jq/issues/2366#issue-1045236954
- https://github.com/itchyny/gojq/issues/62
- https://github.com/wader/jqjq/issues/8

### To-do

- [ ] handle func definitions
- [ ] quickly format jq by appending `fmt` to `jq` on the CLI

### License

This project is licensed under the [MIT License](LICENSE.md).
