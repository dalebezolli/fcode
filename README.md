# wcode - Unf*ck your project folder

> For side-project collectors of all ages and genders.
wcode (which code) provides a simple way to find and navigate to the correct project directory.

As the tool is primarily indended for personal use, I haven't made it simple to install or use, **yet...**

![wcode Showcase](./wcode_showcase.png)

## 🌱 How to install
1. Clone the repo.
2. Set variable WCODE_PATHS with all the paths (space separated) the tool will look for projects
3. Profit?!?

## 🌷 How to use
To use it after compilation run the following command and let the app simplify your life
```sh
wcode; if [ $(echo $?) -eq 0 ]; then cd $(cat ~/.config/wcode/selection); else echo "No project selected"; fi
```
To make your life easier, alias the above command.

## 🧑‍🌾 How to contribute
Feel free to suggest any additions or changes by creating a pull request.
