export default class BoxSelectorController {
  constructor() {
    this.isChecked = this.isChecked.bind(this);
    this.change = this.change.bind(this);
  }

  change(value) {
    this.ngModel = value;
    if (this.onChange) {
      this.onChange(value);
    }
  }

  isChecked(value) {
    return this.ngModel === value;
  }
}
